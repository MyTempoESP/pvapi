---
# pvapi-z9zw
title: Add client_id to checkouts and package_reserved
status: todo
type: task
created_at: 2026-06-04T23:05:00Z
updated_at: 2026-06-04T23:05:00Z
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
