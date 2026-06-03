---
# pvapi-kxod
title: Refactor package_reserved and packages quantity tracking
status: completed
type: feature
priority: normal
created_at: 2026-06-03T01:08:28Z
updated_at: 2026-06-03T01:20:49Z
---

## Problem Statement

After a checkout, the system tracks reserved stock in `package_reserved` and computes remaining availability as `packages.quantity - SUM(package_reserved.quantity)`. This computed availability appears in product catalog views, cart package listing, and checkout stock validation. The dual-purpose role of `package_reserved` — both stock tracker and sales journal — creates complexity: every availability read requires a subquery, and stock can only be derived, not inspected directly.

Additionally, shops and consignees have no visibility into which sold packages were actually picked up (confirmed) vs just purchased.

## Solution

Make `packages.quantity` the live remaining stock — directly decremented at checkout. Keep `package_reserved` as a pure sales journal, gaining a `status` field so sold items can transition from `pending` to `confirmed` when picked up. Store denormalized columns (`shop_id`, `consignment_id`, `shop_value`) on `package_reserved` so sales data survives package deletion.

## User Stories

1. As a shop owner, I want `packages.quantity` to show actual remaining stock, so I can see at a glance how many units are left without computing reservations.
2. As a platform operator, I want sales records to survive even if a consignment package is later deleted, so financial history stays intact.
3. As a consignee, I want to see how many products were sold from each of my consignments, so I can track performance.
4. As a consignee, I want to see how many sold products were confirmed as picked up, so I can distinguish picked-up from still-pending items.
5. As a client, I want checkout to fail cleanly with a clear "insufficient stock" message when another client bought the last unit, so I am not confused by a generic error.
6. As a shop owner removing a product from a consignment, I want previously sold units to not block the removal, since sold stock is tracked separately.
7. As a developer, I want stock decrement and sales recording to happen atomically per package, so no partial state is possible.
8. As a platform operator, I want checkout shop financial breakdowns to be computed from self-contained `package_reserved` data, not dependent on live `packages` rows.

## Implementation Decisions

### Schema changes (single migration)

- **`package_reserved.package_id` FK**: change to `ON DELETE SET NULL`, column becomes nullable. This decouples sales records from package lifecycle.
- **New `package_reserved` columns**:
  - `shop_id INTEGER NOT NULL` — denormalized shop reference, ensures sales attribution survives package deletion
  - `consignment_id INTEGER NOT NULL` — denormalized consignment reference, enables per-consignment sold/confirmed counts without joining through packages
  - `shop_value NUMERIC(12,2) NOT NULL` — price × quantity, for `checkout_shop` breakdown without package join
  - `status TEXT NOT NULL` — lifecycle state (default `pending`, transitions to `confirmed` via external API on pickup)
- **`product_catalog_view`**: replace `SUM(pkg.quantity - SUM(pr.quantity))` with `SUM(pkg.quantity)`. Filter on active consignments retained.
- **`package_decrement_quantity` plpgsql function**: atomically check and decrement `packages.quantity` using `FOR UPDATE` lock, same pattern as existing `product_decrement_inventory`.

### Checkout flow

- **`InsertPackageReservedQuery`**: rewritten as a single CTE per item. The CTE `UPDATE`s `packages SET quantity = quantity - $qty WHERE id = $pid AND quantity >= $qty`, then `INSERT`s into `package_reserved` with all columns including the new denormalized ones. Tax and commission are computed via the same joins as today.
- **`CheckPackageAvailabilityQuery`**: simplified to `SELECT quantity FROM packages WHERE id = $1 FOR UPDATE`. Kept as a pre-check (same transaction) to produce the distinct `ErrInsufficientStock` 409 response. `FOR UPDATE` serializes competing checkouts on the same package so the pre-check is reliable.
- **`InsertCheckoutShopQuery`**: reads `shop_value`, `tax`, `commission` directly from `package_reserved` columns, groups by `shop_id`. No longer joins through `packages` or `consignments`.

### Cart package listing

- **`AvailablePackages`**: drops the `package_reserved` subquery. Filters `pkg.quantity > 0` since sold-out packages are not available.

### Consignee listing

- **`total_products`**: renamed semantically — now means total sold units (`SUM(package_reserved.quantity)` where `package_reserved.consignment_id = consignment.id`), not originally stocked.
- **`sold_products` → `confirmed_products`**: counts `package_reserved.quantity WHERE status = confirmed` for the same consignment.
- **Go structs**: `ConsignmentWithStats.SoldProducts` renamed to `ConfirmedProducts`. `ConsignmentWithStatsResource.ToResource()` maps `confirmedProducts` in JSON.

### Dead code removal

- `CartHoldStock` query and `CartDetails.HoldStock` method — defined but never called.
- `CheckPackageReservedForProductQuery` — only used by `RemoveProductAction.checkNoActiveReservations`, which is being removed.
- `RemoveProductAction.checkNoActiveReservations` guard — removed since `package_reserved` no longer represents held stock.

### Product removal from consignment

- `RemoveProductAction` no longer blocks on active `package_reserved` rows. After removal, `packages.quantity` is incremented back into `products.stock` via `product_increment_inventory` — safe because sold stock is tracked separately in `package_reserved`.

## Testing Decisions

- Existing tests modified to reflect the refactored behavior (view changes, consignee listing JSON shape, checkout stock validation).
- Test principles: test external behavior only (API responses, stock counts in JSON payloads, error codes for insufficient stock), not internal query implementation.
- Prior art: existing integration tests in the repo test checkout flow, consignment listing, and product catalog through the HTTP layer.

## Out of Scope

- The confirmation endpoint that sets `package_reserved.status = "confirmed"`. Only the column and plumbing are added; the API is deferred.
- `checkout_id` FK behavior on `package_reserved` (currently `ON DELETE NO ACTION`). Left as-is for a future decision.
- `ResolvePickupEmailDataQuery` — joins through `packages` for email data. Runs immediately after checkout when packages still exist; deferred until confirmation endpoint design settles.

## Further Notes

- The migration should be a single file following the existing numbering convention.
- `package_reserved.status` values are application-managed; no `CHECK` constraint. `pending` and `confirmed` are the only active values for now.



Triage: ready-for-agent



## Summary of Changes

### Database migration (000037)
- Created `package_decrement_quantity` plpgsql function for atomic check-and-decrement
- Changed `package_reserved.package_id` FK to `ON DELETE SET NULL`, column nullable
- Added denormalized columns: `shop_id`, `consignment_id`, `shop_value`, `status`
- Updated `product_catalog_view`: replaced reservation subquery with direct `SUM(pkg.quantity)`
- Added down migration for full rollback

### Query changes
- **`InsertPackageReservedQuery`**: rewritten as CTE that atomically UPDATEs `packages.quantity` and INSERTs into `package_reserved` with denormalized columns
- **`CheckPackageAvailabilityQuery`**: simplified to `SELECT quantity FROM packages ... FOR UPDATE`
- **`InsertCheckoutShopQuery`**: reads directly from `package_reserved` denormalized columns instead of joining through packages/consignments
- **`AvailablePackages`**: dropped reservation subquery, added `pkg.quantity > 0` filter
- **Consignee listing queries**: `sold_products` → `confirmed_products`, `total_products` now reads from `package_reserved` using denormalized `consignment_id`
- **Dead code**: removed `CartHoldStock` and `CheckPackageReservedForProductQuery`

### Go struct / model changes
- Renamed `ConsignmentWithStats.SoldProducts` → `ConfirmedProducts` (db tag: `confirmed_products`)
- Renamed `ConsignmentWithStatsResource.SoldProducts` → `ConfirmedProducts` (JSON: `confirmedProducts`)
- Removed `CartDetails.HoldStock` method
- Removed `RemoveProductAction.checkNoActiveReservations` and its call in `Exec`

### Test updates
- Updated `checkout_reserved_reflected_in_package_options` to handle sold-out packages (quantity=0) being excluded from package options
- Updated `consignee_list_remessas_happy_path`: `soldProducts` → `confirmedProducts`, `totalProducts` now 0 (no sales yet)
