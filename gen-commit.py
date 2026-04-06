#!/usr/bin/env python3
import subprocess
import os
import sys
from google import genai
import questionary

# Ensure API Key is set
if not os.getenv("GEMINI_API_KEY"):
    print("Error: Please set the GEMINI_API_KEY environment variable.")
    sys.exit(1)

client = genai.Client()

def get_git_status():
    """Parses git status to separate staged and unstaged/untracked files."""
    try:
        status = subprocess.check_output(
            ["git", "status", "--porcelain"],
            stderr=subprocess.STDOUT
        ).decode("utf-8")

        staged = []
        unstaged = []

        for line in status.splitlines():
            if not line:
                continue

            # First two characters represent the status
            state = line[:2]
            # The file path starts at index 3. Strip quotes if git added them due to spaces.
            file_path = line[3:].strip('"')

            # Index 0 is the index (staged) state
            if state[0] in 'MADRC':
                staged.append(file_path)

            # Index 1 is the working tree (unstaged) state, or '??' for untracked
            if state[1] in 'MD' or state == '??':
                unstaged.append(file_path)

        return staged, unstaged
    except subprocess.CalledProcessError:
        return [], []

def get_git_context():
    """Gets the names and diff of currently staged files."""
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
    staged, unstaged = get_git_status()

    # Trigger TUI if nothing is staged
    if not staged:
        if not unstaged:
            print("Working tree clean. Nothing to commit.")
            sys.exit(0)

        print("No files are currently staged for commit.")
        selected_files = questionary.checkbox(
            "Select files to stage (Use Space to select, Enter to confirm):",
            choices=unstaged
        ).ask()

        # If user presses Ctrl+C or selects nothing
        if not selected_files:
            print("No files selected. Aborting.")
            sys.exit(0)

        # Stage the selected files
        subprocess.run(["git", "add"] + selected_files)
        print(f"✅ Staged {len(selected_files)} file(s).")

    # Now that we know things are staged, grab the diff
    files, diff = get_git_context()

    if not diff or not diff.strip():
        # This handles edge cases like staging an empty file
        print("No meaningful changes found in staged files.")
        sys.exit(0)

    print("🤖 Analyzing changes and generating message...")
    message = generate_message(files, diff)

    print("\n--- Suggested Commit Message ---")
    print(message)
    print("--------------------------------\n")

    # TUI for final confirmation
    confirm = questionary.confirm("Apply this commit message?").ask()

    if confirm:
        subprocess.run(["git", "commit", "-m", message])
        print("✅ Committed successfully.")
    else:
        print("❌ Commit aborted (files are still staged).")

if __name__ == "__main__":
    main()
