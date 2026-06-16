---
# pvapi-ndky
title: Cancel checkout endpoint
status: done
type: feature
priority: normal
created_at: 2026-06-16T03:49:59Z
updated_at: 2026-06-16T03:50:02Z
---

## Problem Statement

API clients that confirm checkouts via `POST /checkouts/confirm` have no way to cancel a checkout when a mistake is made or a client changes their mind. Packages marked as RECEIVED (already delivered to the client) cannot be taken back, but PENDING packages can still be cancelled — returning stock, deleting the reservation, and adjusting the shop financial breakdown accordingly.

## Solution

A new `DELETE /checkouts/:unique_id` endpoint, scoped to the same API client that confirmed the checkout. It performs a partial cancellation: PENDING packages are unreserved (stock returned, row deleted), RECEIVED packages are left untouched. `checkout_shop` rows are recomputed or deleted. If no packages remain at all, the checkout itself is deleted.

## User Stories

1. As an API client, I want to cancel a checkout I previously confirmed, so that I can correct mistakes before packages are delivered.
2. As an API client, I want partial cancellation — only cancelling the packages still PENDING — so that already-delivered packages are unaffected.
3. As an API client, I want the checkout_shop financial breakdown to reflect only the remaining RECEIVED packages after a partial cancellation, so the shop's balance stays accurate.
4. As an API client, I want the checkout itself to be deleted when all its packages are cancelled, so there are no orphan rows.
5. As an API client, I want to see how many packages were cancelled and how many remain, so I can confirm the result.
6. As an API client, I want a clear 403 error if I try to cancel another API client's checkout, so authorization is enforced.
7. As an API client, I want a clear 404 error if the unique_id doesn't match any checkout, so I can handle missing checkouts gracefully.
8. As a shop owner, I want cancelled PENDING packages to have their stock returned to the packages table, so the products become available for purchase again.

## Implementation Decisions

### Endpoint contract

- `DELETE /checkouts/:unique_id` — same group, same `apitokenmw.Require` middleware as confirm.
- Request: no body. `:unique_id` path parameter is the checkout's UUID.
- Response (success): `{"success": true, "cancelled_count": N, "remaining_count": M}`
- 403 if `checkouts.api_client_id` does not match the calling API client.
- 404 if no checkout has that unique_id.
- If zero PENDING packages exist → succeed with `cancelled_count: 0`.

### Transaction logic

The entire cancellation runs in a single database transaction with row locks:

1. Lock the `checkouts` row (`SELECT ... FOR UPDATE`). Verify `api_client_id` match (→403). Resolve internal ID.
2. Query all `package_reserved` rows for the checkout. Split into RECEIVED (skip) and PENDING (cancel).
3. For each PENDING row: if `package_id` is not NULL, return stock (`packages.quantity += qty`). Delete the `package_reserved` row.
4. For each `checkout_shop` row: if RECEIVED packages remain for that shop, recompute `shop_value / shop_tax / shop_commission` via SUM over remaining `package_reserved` rows and UPDATE. If zero packages remain for that shop, DELETE the `checkout_shop` row.
5. If zero `checkout_shop` rows remain → DELETE the `checkout` row.

### Modules

- `query/checkout.go` — new queries for: find+lock checkout by unique_id, list package_reserved rows, return stock, delete package_reserved, recompute+update checkout_shop, delete checkout_shop, delete checkout, count remaining checkout_shop rows.
- `models/checkoutmodel/checkout.go` — new `Cancel` function wrapping all business logic in one transaction.
- `controllers/checkout/checkout.go` — new thin handler `CancelHandler`.
- `routes.go` — new route `checkouts.DELETE("/:unique_id", checkout.CancelHandler(pool))`.

### No schema changes

No new columns, tables, or status values. The checkout's status stays `paid` after partial cancellation; RECEIVED packages remain `recebido`.

### Edge cases

- `package_id` is NULL (product deleted from consignment): skip stock return, still delete the `package_reserved` row.
- All packages already RECEIVED: succeed with `cancelled_count: 0`, nothing changes.
- Non-existent unique_id: 404.

## Testing Decisions

Integration tests in the existing style (`api_*_test.go` using `apiHarness`), targeting:

- Full cancellation (all PENDING, zero RECEIVED): checkout deleted, stock returned, checkout_shop rows deleted.
- Partial cancellation (mix of RECEIVED and PENDING): checkout remains, checkout_shop recomputed.
- No-op cancellation (all RECEIVED): cancelled_count=0, nothing changed.
- 403 on mismatched API client.
- 404 on unknown unique_id.

Test harness already provides `confirmCheckout`, `defaultAPISession`, and `doRequest`. Tests follow the pattern of `api_checkout_shop_test.go`.

## Out of Scope

- Cancelling RECEIVED packages (reverse delivery). Not possible.
- Adding a `cancelled` status to checkouts. Not needed; checkout stays `paid` or is deleted.
- Admin-only cancellation. Same-API-client authorization only.
- Cancelling by anything other than `unique_id`.

## Further Notes

Package status transition from `pending` to `recebido` is done manually (offline by consignees). No race condition with automated systems — the transaction is still row-locked for correctness.



Triage: ready-for-agent
