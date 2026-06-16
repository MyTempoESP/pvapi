---
name: test-code
description: "Instructions for running the codebase integration test suite and linter. Use when asked to run, write, or verify tests in this codebase."
---

# Test Code

Instructions for running the codebase integration test suite and linter. Every mention of scripts in this skill is related to the directory containing the SKILL.md.

# Available scripts

- **scripts/run.sh**: Runs Go tests against an ephemeral PostgreSQL test database

## Usage

Do not try to manually run the script steps. Usage:

```bash
scripts/run.sh -r TestNamePattern -v
```

No flags runs `golangci-lint` followed by the full test suite.

| Short | Long | Description |
|-------|------|-------------|
| `-r` | `--run` | Go test `-run` pattern |
| `-v` | `--verbose` | Verbose output |
| `-c` | `--count` | Repeat count (default: 1) |
| `-t` | `--timeout` | Test timeout (default: 120s) |
| — | `--race` | Enable race detector |
| — | `--short` | Skip slow tests |
| — | `--no-lint` | Skip golangci-lint |
| — | `--keep-db` | Don't clean test DB after run |
