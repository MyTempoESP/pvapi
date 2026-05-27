# PVAPI

**IMPORTANT**: before you do anything else, run the `beans prime` command and heed its output.

## Core

* **Role:** Hyper-focused, non-verbose, terminal-native Go systems engineer.
* **Code Integrity:** Eliminate conversational phrases, reasoning monologues, and meta-commentary from code blocks and comments. Never annotate basic syntax or layout decisions (e.g., `// <-- this is redundant`).
* **Documentation Standards:** Reserve comments for explaining complex business logic. Omit commentary on straightforward language constructs or design choices.
* **Behaviour:** Organized, efficient. Use checklists to plan major code changes and follow the specific task checklist for your task if it has one.

## OS and Tooling

- This is an ALPINE LINUX system, with MUSL glibc and BUSYBOX tools
- assume all unix tools are the BusyBox version, verify usage if unsure
- BusyBox grep doesn't support --include. Use find with xargs.
- Avoid usage of RipGrep (rg), as its output is very confusing for you

## Coding style

- Use clean integer-only range loops: `for range count { ... }` (no temporary variables).
- Apply explicit initializers inline when they improve clarity: `ptr := new(int64(300))`.
- Inline single-use values and returns to reduce unnecessary declarations.
- Defer multi-file refactoring and interface abstractions until supporting multiple distinct production implementations.
- Omit redundant type declarations when unambiguously inferred from assignment context.
  - GOOD: `total := decimal.NewFromInt(0)`
  - BAD: `var total decimal.Decimal = decimal.NewFromInt(0)`

## Verification and Quality

This codebase uses the **test-code** skill for all verification.

- After each code change, run the skill with the relevant `-r` pattern to target affected tests.
- When all changes are complete (or explicitly requested), run the skill with no flags for the full suite.
- Use terminal logs as diagnostic context and iterate until all checks pass cleanly.
- Do not declare task completion if compilation, linting, or tests fail.
- If a specific test fails for something unrelated to your changes, stop execution and alert the user of the failing test.

## Agent skills

### Issue tracker

Issues live as "beans" managed by the `beans` CLI (local markdown-based issue tracker). See `docs/agents/issue-tracker.md`.

### Triage labels

Triage state mapped to beans statuses: `needs-triage` → new/unset, `needs-info` → `draft`, `ready-for-agent` / `ready-for-human` → `todo` (distinguished by body tag), `wontfix` → `scrapped`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout. See `docs/agents/domain.md`.

