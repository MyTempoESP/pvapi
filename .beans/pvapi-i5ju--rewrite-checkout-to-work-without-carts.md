---
# pvapi-i5ju
title: Rewrite checkout to work without carts
status: completed
type: task
priority: normal
created_at: 2026-06-04T23:05:11Z
updated_at: 2026-06-04T23:22:28Z
blocked_by:
    - pvapi-z9zw
---

## What to build

Remove all cart dependencies from the checkout flow. The checkout currently requires a cart to exist (resolveCart gate, FinishCart marking). After this slice, the checkout endpoint works purely with session_id + external_reference + package items — no cart involvement.

Changes:
- Remove `resolveCart` (cart-existence + item-count validation gate)
- Remove `FinishCartQuery` (FINISHED- session_id marking)
- Rewrite `checkIdempotency` to use `external_reference` alone (no cart session join)
- `writeCheckout` returns `checkout_id` alongside `uniqueID`
- `writeCheckout` writes `client_id` to `checkouts` and `package_reserved` (alongside existing `checkout_shop`)
- `ConfirmResponse` drops `CartID`, gains `CheckoutID`
- `CreateAndAssociate` and `SetCheckoutShopClientID` are replaced with direct writes using `checkout_id`
- `sendConfirmationAfterCheckout` writes client_id to checkouts, checkout_shop, and package_reserved
- Update all checkout tests (api_cart_checkout_test.go, api_checkout_client_test.go, api_checkout_shop_test.go) to not use createCartItem — they POST to /checkouts/confirm directly with package IDs

## Acceptance criteria

- [ ] Checkout succeeds without any cart rows in the database
- [ ] Idempotent retry with same external_reference returns same unique_id (no cart needed)
- [ ] Client email in checkout creates client and writes client_id to checkouts, checkout_shop, and package_reserved
- [ ] All checkout-related tests pass (TestCartAndCheckout, checkout client tests, checkout shop tests)
- [ ] No reference to carts or cartmodel remains in checkout or client code paths

## Blocked by

- pvapi-z9zw
