#!/usr/bin/env python3
import subprocess
import os
import sys
import questionary
from google import genai
from google.genai import errors

# Import the necessary tenacity modules
from tenacity import retry, wait_random_exponential, stop_after_attempt, retry_if_exception

if not os.getenv("GEMINI_API_KEY"):
    print("Error: Please set the GEMINI_API_KEY environment variable.")
    sys.exit(1)

client = genai.Client()

def should_retry_api_error(exception):
    """
    Evaluates if the exception is a recoverable HTTP error.
    Returns True for 429 (Rate Limit) and 503 (Service Unavailable).
    """
    if isinstance(exception, errors.ClientError):
        # Extract the status code from the exception arguments
        status_code = exception.args[0] if len(exception.args) > 0 else None

        if status_code in (429, 503):
            # We print a subtle log here just so the user knows the CLI isn't hanging
            print("⏳ Rate limit or server load detected. Backing off and retrying...")
            return True

    return False

# The Tenacity Decorator
@retry(
    wait=wait_random_exponential(multiplier=2, max=60),
    stop=stop_after_attempt(5),
    retry=retry_if_exception(should_retry_api_error),
    reraise=True
)
def generate_message(files, diff):
    """Generates the commit message. Network retries are handled externally by Tenacity."""

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

    response = client.models.generate_content(
        model='gemini-2.5-flash',
        contents=prompt
    )

    return response.text.strip()


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

    print("🤖 Analyzing changes and generating message...")

    try:
        message = generate_message(files, diff)
    except errors.ClientError as e:
        # If we get here, it means we either hit a non-429 error (like a 400 Bad Request),
        # OR we exhausted all 5 of our retries on a 429.
        print(f"\n❌ API Error: {e}")
        print("Please check your network connection or API quota.")
        sys.exit(1)

    print("\n--- Suggested Commit Message ---")
    print(message)
    print("--------------------------------\n")

    confirm = questionary.confirm("Apply this commit message?").ask()

    if confirm:
        subprocess.run(["git", "commit", "-m", message])
        print("✅ Committed successfully.")
    else:
        print("❌ Commit aborted (files are still staged).")

if __name__ == "__main__":
    main()
