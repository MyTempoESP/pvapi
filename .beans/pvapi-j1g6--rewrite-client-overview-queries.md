---
# pvapi-j1g6
title: Rewrite client overview queries
status: completed
type: task
priority: normal
created_at: 2026-06-04T23:05:32Z
updated_at: 2026-06-05T00:02:12Z
blocked_by:
    - pvapi-z9zw
---

## What to build

Rewrite client overview queries to use `package_reserved.client_id` instead of joining through `carts.client_id` (since carts are being deleted).

Changes:
- `ClientOverviewCountsQuery`: replace `JOIN carts c ON c.id = chk.cart_id WHERE c.client_id = $1` with `WHERE pr.client_id = $1`
- `ClientOverviewPackagesQuery`: replace `JOIN carts c ON c.id = chk.cart_id WHERE c.client_id = $1` with `WHERE pr.client_id = $1`
- Update `api_client_overview_test.go` if needed

## Acceptance criteria

- [ ] Client overview endpoint returns correct order counts and packages
- [ ] No reference to `carts` table in overview queries
- [ ] Client overview test passes

## Blocked by

- pvapi-z9zw


## Summary of Changes

Rewrote `ClientOverviewCountsQuery` and `ClientOverviewPackagesQuery` in `query/client_overview.go` to use `pr.client_id` instead of joining through `checkouts.client_id`.

Changes:
- ****: Removed  entirely. Now filters directly with .
- ****: Replaced  with . Kept the JOIN to  for .
- No test changes needed — all 5  subtests pass.

**Acceptance criteria verified:**
- [x] Client overview endpoint returns correct order counts and packages
- [x] No reference to `carts` table in overview queries
- [x] Client overview test passes
