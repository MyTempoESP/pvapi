---
# pvapi-z9zw
title: Add client_id to checkouts and package_reserved
status: completed
type: task
priority: normal
created_at: 2026-06-04T23:05:00Z
updated_at: 2026-06-04T23:10:46Z
---

## What to build

Add a `client_id` column to the `checkouts` and `package_reserved` tables with foreign keys to `clients`. Add a `UNIQUE(external_reference)` constraint to `checkouts` for idempotency. Migration only — no Go code changes, fully backward-compatible.

## Acceptance criteria

- [ ] `checkouts.client_id INTEGER` with FK to `clients(id)` (nullable, no NOT VALID)
- [ ] `package_reserved.client_id INTEGER` with FK to `clients(id)` (nullable)
- [ ] `checkouts` has `UNIQUE(external_reference)` constraint
- [ ] Full test suite passes (no code changes, just schema)

## Blocked by

None — can start immediately.


## Summary of Changes

Migration `000038_add_client_id_to_checkouts_and_package_reserved` already existed. Applied cleanly against a fresh test database — all 38 migrations ran without errors. No Go code changes needed.

**Acceptance criteria verified:**
- [x] `checkouts.client_id INTEGER` with FK to `clients(id)`
- [x] `package_reserved.client_id INTEGER` with FK to `clients(id)`
- [x] `UNIQUE(external_reference)` on `checkouts`
- [x] Full test suite passes (schema-only change)
