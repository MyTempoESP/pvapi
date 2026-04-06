#!/usr/bin/env python3
import subprocess
import os
from google import genai

# The SDK automatically looks for the GEMINI_API_KEY environment variable.
if not os.getenv("GEMINI_API_KEY"):
    print("Error: Please set the GEMINI_API_KEY environment variable.")
    exit(1)

# Initialize the modern client
client = genai.Client()

def get_git_context():
    try:
        # Get list of staged files
        files = subprocess.check_output(
            ["git", "diff", "--cached", "--name-only"],
            stderr=subprocess.STDOUT
        ).decode("utf-8")

        # Get the actual diff content
        diff = subprocess.check_output(
            ["git", "diff", "--cached"],
            stderr=subprocess.STDOUT
        ).decode("utf-8")

        return files, diff
    except subprocess.CalledProcessError:
        return None, None

def generate_message(files, diff):
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
    - If changes are significant, include a short bulleted body.
    - Output ONLY the commit message. No markdown blocks, no chat.
    """

    response = client.models.generate_content(
        model='gemini-2.5-flash',
        contents=prompt
    )

    return response.text.strip()

def main():
    files, diff = get_git_context()

    if not diff or not diff.strip():
        print("No staged changes found. Use 'git add' first.")
        return

    print("🤖 Analyzing changes and generating message...")
    message = generate_message(files, diff)

    print("\n--- Suggested Commit Message ---")
    print(message)
    print("--------------------------------\n")

    confirm = input("Apply this commit message? (y/n): ")
    if confirm.lower() == 'y':
        subprocess.run(["git", "commit", "-m", message])
        print("✅ Committed successfully.")
    else:
        print("❌ Commit aborted.")

if __name__ == "__main__":
    main()
