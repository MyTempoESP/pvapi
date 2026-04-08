#!/usr/bin/env python3
import subprocess
import os
import sys
import questionary
from google import genai
from openai import OpenAI
import openai

# Import the necessary tenacity modules for DeepSeek
from tenacity import retry, wait_random_exponential, stop_after_attempt, retry_if_exception

# Check API Keys upfront
GEMINI_KEY = os.getenv("GEMINI_API_KEY")
DEEPSEEK_KEY = os.getenv("DEEPSEEK_API_KEY")

if not GEMINI_KEY or not DEEPSEEK_KEY:
    print("⚠️  Warning: For the failover system to work, please set BOTH environment variables:")
    print("   export GEMINI_API_KEY='your_key'")
    print("   export DEEPSEEK_API_KEY='your_key'")
    sys.exit(1)

# Initialize DeepSeek Client
ds_client = OpenAI(api_key=DEEPSEEK_KEY, base_url="https://api.deepseek.com")

def should_retry_deepseek_error(exception):
    """Only retry DeepSeek if it throws a Rate Limit or Server Error."""
    if isinstance(exception, (openai.RateLimitError, openai.InternalServerError, openai.APIConnectionError)):
        print("⏳ DeepSeek rate limit or server load detected. Backing off and retrying...")
        return True
    return False

@retry(
    wait=wait_random_exponential(multiplier=2, max=60),
    stop=stop_after_attempt(5),
    retry=retry_if_exception(should_retry_deepseek_error),
    reraise=True
)
def generate_with_deepseek(prompt):
    """The Fallback: Uses DeepSeek with Tenacity retries."""
    response = ds_client.chat.completions.create(
        model="deepseek-chat",
        messages=[{"role": "user", "content": prompt}],
        temperature=0.2,
        max_tokens=200
    )
    return response.choices[0].message.content.strip(), "DeepSeek V3"

def generate_message(files, diff):
    """The Router: Tries Gemini first, fast-fails to DeepSeek."""

    prompt = f"""
    You are an expert sysadmin and developer. Generate a Git commit message following the
    Conventional Commits specification (semantic commits) based on the diff below.

    Files changed:
    {files}

    Diff:
    {diff}

    Guidelines:
    - Use format: <type>(<scope>): <description>
    - Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
    - Description should be in imperative, present tense (e.g., "add" not "added").
    - If changes are significant, include a short bulleted body, using markdown dash syntax.
    - Output ONLY the commit message. No markdown blocks, no chat.
    """

    print("🤖 Attempting generation with Gemini 2.5 Flash...")

    try:
        # We instantiate the client here. The SDK automatically picks up GEMINI_API_KEY.
        gemini_client = genai.Client()
        response = gemini_client.models.generate_content(
            model='gemini-2.5-flash',
            contents=prompt
        )
        return response.text.strip(), "Gemini 2.5 Flash"

    except Exception as e:
        # We catch ANY error from Gemini (429, 503, Auth, etc.) and immediately failover.
        # We extract the status code cleanly if it's an APIError, or just print the raw error.
        error_msg = getattr(e, 'status_code', str(e).split('.')[0])
        print(f"⚠️  Gemini failed ({error_msg}). Instantly failing over to DeepSeek...")

        # Call the DeepSeek function, which has its own built-in retry logic
        return generate_with_deepseek(prompt)


def get_git_status():
    try:
        status = subprocess.check_output(
            ["git", "status", "--porcelain"],
            stderr=subprocess.STDOUT
        ).decode("utf-8")

        staged, unstaged = [], []

        for line in status.splitlines():
            if not line: continue
            state, file_path = line[:2], line[3:].strip('"')
            if state[0] in 'MADRC': staged.append(file_path)
            if state[1] in 'MD' or state == '??': unstaged.append(file_path)

        return staged, unstaged
    except subprocess.CalledProcessError:
        return [], []

def get_git_context():
    try:
        files = subprocess.check_output(
            ["git", "diff", "--cached", "--name-only"],
            stderr=subprocess.STDOUT
        ).decode("utf-8")

        diff = subprocess.check_output(
            ["git", "diff", "--cached"],
            stderr=subprocess.STDOUT
        ).decode("utf-8")

        return files, diff
    except subprocess.CalledProcessError:
        return None, None

def main():
    staged, unstaged = get_git_status()

    if not staged:
        if not unstaged:
            print("Working tree clean. Nothing to commit.")
            sys.exit(0)

        print("No files are currently staged for commit.")
        selected_files = questionary.checkbox(
            "Select files to stage (Use Space to select, Enter to confirm):",
            choices=unstaged
        ).ask()

        if not selected_files:
            print("No files selected. Aborting.")
            sys.exit(0)

        subprocess.run(["git", "add"] + selected_files)
        print(f"✅ Staged {len(selected_files)} file(s).")

    files, diff = get_git_context()

    if not diff or not diff.strip():
        print("No meaningful changes found in staged files.")
        sys.exit(0)

    try:
        message, model_used = generate_message(files, diff)
    except Exception as e:
        print(f"\n❌ CRITICAL: Both APIs failed. Last error from DeepSeek: {e}")
        sys.exit(1)

    print(f"\n--- Suggested Commit Message (Powered by {model_used}) ---")
    print(message)
    print("----------------------------------------------------------\n")

    confirm = questionary.confirm("Apply this commit message?").ask()

    if confirm:
        subprocess.run(["git", "commit", "-m", message])
        print("✅ Committed successfully.")
    else:
        print("❌ Commit aborted (files are still staged).")

if __name__ == "__main__":
    main()
