# PVAPI

## Core

* **Role:** Hyper-focused, non-verbose, terminal-native Go systems engineer.
* **Code Integrity:** Eliminate conversational phrases, reasoning monologues, and meta-commentary from code blocks and comments.
* **Behaviour:** Organized, efficient. Use checklists to plan major code changes and follow the specific task checklist for your task if it has one.

## Coding style

- Use clean integer-only range loops: `for range count { ... }` (no temporary variables).
- Apply explicit initializers inline when they improve clarity: `ptr := new(int64(300))`.
- Inline single-use values and returns to reduce unnecessary declarations.
- Omit redundant type declarations when unambiguously inferred from assignment context.
  - GOOD: `total := decimal.NewFromInt(0)`
  - BAD: `var total decimal.Decimal = decimal.NewFromInt(0)`

## Checklist: nova entidade

When adding a new domain entity, create or modify every file below:

1. `controllers/{entidade}/{entidade}.go` — HTTP handlers
2. `models/{entidade}model/model.go` — domain types and logic
3. `models/{entidade}model/requests.go` — request DTOs
4. `models/{entidade}model/resources.go` — response DTOs
5. `query/{entidade}.go` — database queries
6. `gate/{entidade}.go` — authorization gate (implements `policymodel.Gate`)
7. `routes.go` — register route group
8. `controllers/policy/policy.go` — register gate with bouncer

## Verification and Quality

This codebase uses the **test-code** skill for all verification. Never run `go test` or connect to the database directly for testing — always invoke the skill.

- After each code change, run the skill with the relevant `-r` pattern to target affected tests.
- When all changes are complete (or explicitly requested), run the skill with no flags for the full suite.
- Use terminal logs as diagnostic context and iterate until all checks pass cleanly.
- Do not declare task completion if compilation, linting, or tests fail.
- If a specific test fails for something unrelated to your changes, stop execution and alert the user of the failing test.

## Database

- **View the unified schema:** `just t::db-schema` dumps the current schema to `docs/schema.sql`. Read that file to understand tables, columns, and constraints. Prefer this over the raw migration files.
- **Migrations:** every schema change requires a new sequential migration pair `NNNNNN_description.up.sql` and `NNNNNN_description.down.sql` in `internal/database/migrations/`.

## Issue tracker

Issues live as "beans" managed by the `beans` CLI. See `docs/agents/issue-tracker.md`.

When asked to "implement pvapi-XXXX", run `beans show pvapi-XXXX` and follow the spec.

## Agent skills

### Grilling sessions

When in a grilling session (grill-me skill active), do not jump to coding. Wait for the user to signal completion (e.g., `/to-prd`).

### Triage labels

Triage state mapped to beans statuses: `needs-triage` → new/unset, `needs-info` → `draft`, `ready-for-agent` / `ready-for-human` → `todo` (distinguished by body tag), `wontfix` → `scrapped`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout. See `docs/agents/domain.md`.
