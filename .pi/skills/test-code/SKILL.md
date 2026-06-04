---
name: test-code
description: "Runs Go tests against a temporary PostgreSQL database. Manages the full test lifecycle: start test DB, run targeted or full test suites, lint, close DB. Use when asked to run, write, or verify tests in this codebase."
---

# Test Code

Runs Go tests against an ephemeral PostgreSQL test database managed by `just`.

## Usage

End-to-end wrapper around `just t::test-db`, `golangci-lint`, and `go test`:

```bash
SKILL_DIR="$(dirname "$(readlink -f "$0")")"
"$SKILL_DIR/scripts/run.sh" -r TestNamePattern -v
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
