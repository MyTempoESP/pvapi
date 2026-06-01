---
# pvapi-20qk
title: Freeze shop-level values at checkout
status: completed
type: feature
priority: normal
created_at: 2026-06-01T09:34:37Z
updated_at: 2026-06-01T09:44:28Z
---

## Problem Statement

When a client completes a checkout, the checkout's `value_paid` is the amount the client typed in — not necessarily what the products actually cost. Shop finances currently display this user-input value as the checkout total, which means a shop's reported earnings can diverge from the real calculated prices of the packages they sold. Tax and commission are already computed correctly at the package level, but the base value shown to the shop is wrong.

The platform needs to freeze the real calculated shop value, tax, and commission at checkout time so that shop financial reports reflect actual package prices, not user-supplied payment amounts.

## Solution

On checkout creation, compute and persist per-shop financial breakdowns in a new `checkout_shop` table. Each row captures the shop's actual earnings from that checkout — `shop_value` (sum of package prices), `shop_tax` (platform tax), and `shop_commission` (consignment commission) — all as frozen monetary values calculated at checkout time.

The shop finances listing switches its data source from the user-input `checkouts.value` to `checkout_shop.shop_value`. The discrepancy between what the client typed and what the packages cost is captured as `value_discrepancy` on the checkout for audit purposes.

## User Stories

1. As a shop owner viewing my finances, I want to see the actual calculated value of the packages sold in each checkout, so that my reported earnings match the real product prices rather than whatever amount the client typed.

2. As a shop owner viewing my finances, I want my tax and commission deductions to be computed from the actual package values, so that my net earnings are correct.

3. As a platform operator, I want the tax and commission for each checkout to be frozen at the moment of purchase, so that later changes to tax rates or commission structures do not retroactively affect settled transactions.

4. As a platform operator, I want to detect when a client's stated payment amount differs from the actual package prices, so that I can audit potential discrepancies or payment errors.

5. As a developer maintaining the checkout flow, I want the shop-level financial freeze to happen atomically within the same transaction as checkout creation, so that partial failures cannot produce inconsistent financial data.

6. As a shop owner, I want my available balance calculation to use the frozen shop values, so that my withdrawable amount reflects real earnings.

7. As a platform operator viewing a multi-shop checkout, I want each shop's financials separated into their own rows, so that a single checkout spanning products from different shops correctly attributes value to each shop.

## Implementation Decisions

### Schema changes

- New table `checkout_shop` with columns: `checkout_id` (FK → checkouts.id), `shop_id` (FK → shops.id), `checkout_unique_id` (uuid, denormalized), `shop_value` (numeric), `shop_tax` (numeric), `shop_commission` (numeric), `created_at` (timestamptz). Primary key is the composite `(checkout_id, shop_id)`.
- New column `value_discrepancy numeric(12,2)` on the `checkouts` table, computed as `value_paid - SUM(checkout_shop.shop_value)`.

### Checkout freeze logic

- `shop_value` is `SUM(pkg.price × quantity)` per shop, derived from the confirmed packages, not from cart_product frozen prices or the base product price. Checkout operates on packages — the package price is the atomic unit of the transaction.
- `shop_tax` and `shop_commission` use the identical formulas already in `package_reserved` (`platform_cut` percentage and consignment `PERCENT`/`FIXED` commission logic), aggregated per shop. The formulas are duplicated in the new query rather than extracted — the two queries serve different purposes (per-package inserts vs per-shop aggregates) and rarely change.
- Aggregation groups by `shop_id` via the JOIN chain: ConfirmItem → packages → consignments → shops.
- All computation happens inside the existing `writeCheckout` transaction, after the checkout row is inserted but before `package_reserved` rows. This ensures atomicity without reading back data that was just written.

### Query rewrites

- `ShopCheckoutsQuery`: switches from the multi-JOIN through `package_reserved` → packages → consignments to a direct query on `checkout_shop` joined with `checkouts` (for `unique_id`, `description`, `processed_at` ordering).
- `CheckoutEarningsByShopQuery`: switches from `SUM(chk.value - pr.tax - pr.commission)` to `SUM(cs.shop_value - cs.shop_tax - cs.shop_commission)` from `checkout_shop`.
- `AvailableBalance`: same logic, different data source. Still locks the shop row for concurrency control.

### Existing data

- `package_reserved.tax` and `package_reserved.commission` are preserved as-is. The `checkout_shop` table duplicates the aggregated values at the shop level but does not replace the per-package detail.
- Backfill of existing checkouts into `checkout_shop` is out of scope for this PRD.

### Shop finances display

- The shop's checkout report shows `shop_value` as the checkout value. The user-input `checkouts.value` is not exposed to the shop. The shop never sees the `value_paid` field — only the calculated `shop_value`.

## Testing Decisions

- What makes a good test: only test external behavior — given a cart with products from known shops, when checkout is confirmed, verify that `checkout_shop` rows contain the correct `shop_value`, `shop_tax`, and `shop_commission` values, and that `value_discrepancy` is correctly computed. Verify the shop finances endpoint returns `shop_value` not `value_paid`.
- Modules to test: the checkout freeze logic (module 2), the shop finances endpoint (end-to-end via the API test pattern already in the codebase).
- Prior art: the API integration tests in `api_cart_checkout_test.go` and `api_checkout_client_test.go` exercise the full checkout flow against a test database and verify response payloads. Follow the same pattern.

## Out of Scope

- Backfilling `checkout_shop` rows for existing checkouts. Historical data will not appear in the new queries until a separate backfill migration is done.
- Adding a `subtotal` column to `package_reserved` to persist the per-package subtotal.
- Removing or altering the `package_reserved.tax` / `package_reserved.commission` columns.
- Enforcing that `value_paid` must match the calculated shop values. The discrepancy is recorded but not rejected.
- Any changes to the `CartHoldStock` query or the `cart_product.price` column.

## Further Notes

- A single checkout containing packages from multiple shops produces one `checkout_shop` row per shop. Each shop only sees its own row in their financial report.
- The tax and commission formulas in `checkout_shop` are a deliberate duplication of the formulas in `InsertPackageReservedQuery`. Both queries use the same inputs (package price, quantity, platform cut, commission type/value) and produce identical results. Extracting shared SQL fragments is deferred as overengineering for two queries that change extremely rarely.



Triage: ready-for-agent



## Summary of Changes

### Schema
- New table `checkout_shop` (migration 000035): stores per-shop frozen financial breakdown at checkout time. Columns: `checkout_id`, `shop_id`, `checkout_unique_id`, `shop_value`, `shop_tax`, `shop_commission`, `created_at`. Composite PK (`checkout_id`, `shop_id`).
- New column `value_discrepancy numeric(12,2)` on `checkouts`: computed as `value_paid - SUM(shop_value)`.

### Checkout freeze logic
- `writeCheckout` in `models/checkoutmodel/checkout.go`: after inserting `package_reserved` rows, inserts per-shop aggregates into `checkout_shop` via `InsertCheckoutShopQuery`, then computes `value_discrepancy` via `UpdateCheckoutDiscrepancyQuery`. All inside the same transaction.
- `InsertCheckoutShopQuery`: aggregates from `package_reserved` → `packages` → `products` → `consignments`, computing `shop_value`, `shop_tax`, and `shop_commission` using the same formula as `InsertPackageReservedQuery`.

### Query rewrites
- `ShopCheckoutsQuery`: switched from JOIN through `package_reserved` to direct query on `checkout_shop` (`cs.shop_value` instead of `chk.value`).
- `CheckoutEarningsByShopQuery`: switched from `chk.value - pr.tax - pr.commission` to `cs.shop_value - cs.shop_tax - cs.shop_commission`.

### Tests
- New `TestCheckoutShopFreeze` in `api_checkout_shop_test.go`: verifies `checkout_shop` rows (FIXED and PERCENT commission), `value_discrepancy` computation, shop checkout report returns `shop_value`, totals use `shop_value`, available balance uses `checkout_shop`.
