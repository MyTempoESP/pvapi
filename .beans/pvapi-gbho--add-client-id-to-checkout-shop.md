---
# pvapi-gbho
title: Add client_id to checkout_shop
status: todo
type: feature
priority: normal
created_at: 2026-06-01T19:04:31Z
updated_at: 2026-06-01T19:04:34Z
---

## Problem Statement

When a client completes a checkout, their identity is associated to the cart via `carts.client_id`, but not to the `checkout_shop` rows that represent the per-shop financial breakdown of that checkout. This means any query that works from `checkout_shop` has no direct path to the client — it must join through `checkouts → carts → clients`, which is fragile and indirect.

## Solution

Add a nullable `client_id` column to `checkout_shop`, populated at the same time as `carts.client_id` (during post-checkout client creation/association). The column carries a foreign key to `clients(user_id)`, matching the existing pattern on `carts`.

## User Stories

1. As a system, I want each checkout_shop row to carry a client_id, so that per-shop earnings can be directly attributed to the purchasing client without multi-hop joins.
2. As a system, I want the client_id on checkout_shop populated atomically with the cart's client association, so that data consistency between cart and checkout_shop is guaranteed.
3. As a system, I want the column to be nullable, so that checkouts without a known client (no email provided) are represented correctly as NULL.
4. As a system, I want a foreign key constraint to clients(user_id), so that referential integrity is enforced at the database level.
5. As a developer, I want the populate logic to reuse the existing CreateAndAssociate transaction boundary, so that no new transaction orchestration is needed.

## Implementation Decisions

- **Schema**: New migration adding `client_id integer` to `checkout_shop` with a foreign key referencing `clients(user_id)`. Nullable. Matches the cart pattern.
- **Query**: The UPDATE uses a JOIN instead of a subquery:
  ```sql
  UPDATE checkout_shop cs
  SET client_id = $1
  FROM checkouts chk
  WHERE chk.cart_id = $2
    AND cs.checkout_id = chk.id
  ```
- **New function**: `clientmodel.SetCheckoutShopClientID(ctx, tx, clientID, cartID)` — executes the join-based UPDATE. Lives in the clientmodel package.
- **Wiring**: Called immediately after `SetCartClientID` inside `CreateAndAssociate`, within the same transaction. Both succeed or both roll back.
- **Existing queries and endpoints are unchanged** — `ShopCheckoutsQuery`, `CheckoutEarningsByShopQuery`, and the shop report endpoint do not expose client_id for now. That is a separate feature.

## Testing Decisions

- Integration tests (API-level) are the right fit: this is a write-path change that piggybacks on the existing checkout → client association flow.
- Prior art: `api_checkout_shop_test.go` already queries `checkout_shop` directly with raw SQL after checkout to assert column values.
- A good test: create a checkout with `client_email` set, then verify the `checkout_shop` row has the correct `client_id` populated. Verify NULL for a checkout without `client_email`.

## Out of Scope

- Exposing `client_id` in shop checkout report responses (`/estabelecimentos/checkouts`).
- Backfilling `client_id` for historical checkout_shop rows.
- Adding `client_id` to any existing read-side queries or views.

## Further Notes

The cart FK uses `REFERENCES public.clients(user_id)` — this migration should do the same for consistency. The cart FK was created inline in the init migration; the checkout_shop FK should follow the same column-as-reference pattern.

\n\nTriage: ready-for-agent
