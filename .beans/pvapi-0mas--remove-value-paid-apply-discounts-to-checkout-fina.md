---
# pvapi-0mas
title: Remove value_paid, apply discounts to checkout financials
status: todo
type: feature
priority: normal
created_at: 2026-06-11T01:10:26Z
updated_at: 2026-06-11T01:10:29Z
---

## Problem Statement

API integrators are blocked by the required `value_paid` field on `POST /checkouts/confirm`. They don't know what value to send because the actual total depends on server-side package pricing. This field serves no current business purpose — the real financial truth lives in `checkout_shop.shop_value` (derived from actual package prices).

Additionally, `packages.discount` is stored and exposed in the catalog API but never applied during checkout. The financial computations for `shop_value`, `tax`, and `commission` use the undiscounted `packages.price`, meaning checkout financials don't reflect discounts at all.

## Solution

1. Remove `value_paid` from the `POST /checkouts/confirm` API contract.
2. Auto-compute `checkouts.value` on the server from actual package pricing with discounts applied.
3. Apply `packages.discount` (a percentage, 0–100) to `shop_value`, `tax`, and `commission` computations during checkout reservation.
4. Drop the `checkouts.value` and `checkouts.value_discrepancy` columns — they're redundant once the value is auto-computed and the discrepancy is always zero.
5. Add a DB constraint and API validation to bound `packages.discount` to [0, 100].

## User Stories

1. As an API integrator, I want to POST to /checkouts/confirm without sending a value_paid field, so that I don't have to pre-compute or guess what the checkout total should be.
2. As an API integrator, I want the checkout value to be computed server-side from actual package prices with discounts applied, so that the financial records are accurate and consistent.
3. As a shop owner, I want discounts on my consignment products to reduce the shop_value recorded in checkout_shop, so that my financial reports reflect the actual amount the client paid for my products.
4. As a shop owner, I want tax and commission calculations to use the discounted price, so that I'm not overcharged on platform fees when I offer discounts.
5. As a platform operator, I want to prevent invalid discount values (negative or >100%), so that the financial pipeline doesn't encounter nonsensical prices.
6. As a developer, I want the checkouts table to only contain data that has a consumer, so that the schema doesn't accumulate write-only dead columns.

## Implementation Decisions

### API contract change

The `value_paid` field is removed from the `POST /checkouts/confirm` JSON body. Callers no longer send this field. The `ConfirmRequest` struct drops the `ValuePaid` field.

### Checkout value computation

The checkout value is pre-computed before insertion. During stock validation (which already does `SELECT ... FOR UPDATE` on packages), the query is extended to also return `packages.price` and `packages.discount`. The effective price per package is computed in Go as `price × (1 - discount/100)`, then summed across all items to produce the checkout total. This total is passed to the checkout insert query.

### Discount application in package_reserved

The `InsertPackageReservedQuery` is updated. The `packages` row returned by the `UPDATE ... RETURNING` now includes `discount`. The effective price is `d.price × (1 - d.discount / 100.0)`. This effective price is used for:

- `shop_value`: `effective_price × quantity`
- `tax` (platform_cut): `(p.platform_cut / 100.0) × effective_price × quantity`
- `commission`: `(c.comission_value / 100.0) × effective_price × quantity` (percent type) or `c.comission_value` (fixed type, unchanged)

### Schema changes

A single migration drops two columns from `checkouts`:
- `value` — no longer needed; financial truth is in `checkout_shop.shop_value`
- `value_discrepancy` — no longer meaningful; was `value - SUM(shop_value)`, now always zero

The migration also adds a check constraint on `packages`:
- `ALTER TABLE packages ADD CONSTRAINT discount_range CHECK (discount >= 0 AND discount <= 100)`

### Discount validation at API layer

The `AddProductAction.Exec` method validates that each product's `discount` is between 0 and 100 (inclusive) before inserting. This catches errors early with a clear error message.

## Testing Decisions

### What makes a good test

Tests should verify external API behavior: the request body shape, the response, and the resulting database state. Do not test internal function signatures or implementation details.

### Test patterns

Follow existing patterns in the codebase:
- `api_checkout_shop_test.go` — verifies `checkout_shop` rows and financial computations after checkout
- `api_checkout_client_test.go` — verifies checkout confirmation with client identity
- `api_consignment_test.go` — verifies consignment product creation

### Test updates required

1. All checkout confirmation test fixtures must remove `"value_paid"` from request bodies.
2. `api_checkout_shop_test.go` — update expected `shop_value`, `shop_tax`, `shop_commission` values to reflect the discount being applied. The existing test at `value_discrepancy_computed` is removed (column dropped). The `shop_checkout_report_uses_shop_value` assertions about `value_paid` and `value_discrepancy` not being in the response are removed.
3. `api_consignment_test.go` — add test for discount boundary validation (discount=101 rejected, discount=-1 rejected).
4. `api_fixtures_test.go` — update `confirmCheckout` helper signature to drop `valuePaid` parameter.

## Out of Scope

- Gift cards, vouchers, or other mechanisms that would create a non-zero discrepancy between payment and package totals
- Discount application at the cart/browsing level (only affects checkout reservation)
- Product-level `discount` field (separate from package-level discount)
- Any API endpoint that currently returns `checkouts.value` (none do; verified)

## Further Notes

- The `checkouts.value` column has zero readers in application code and zero API exposure. The shop checkout report reads from `checkout_shop.shop_value`, and the client overview reads from `package_reserved.shop_value`.
- The `packages.discount` column already exists, is populated at consignment product creation, and is exposed in catalog API responses. It was simply never applied to financial computations.

Triage: ready-for-agent
