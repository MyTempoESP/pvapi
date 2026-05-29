---
# pvapi-7vz8
title: Client account auto-creation during checkout
status: completed
type: feature
priority: normal
created_at: 2026-05-29T23:12:36Z
updated_at: 2026-05-29T23:24:47Z
blocking:
    - pvapi-63xz
---

## Problem Statement

When a buyer completes a checkout on the PVAuto storefront, the back-end has no way to create a client account for them automatically. Currently, client accounts are only created by shop operators for walk-in customers (via the consignment flow using `CreateForCart`). This means returning buyers have no persistent identity tied to their purchases, and there is no workflow to follow up with them post-purchase.

## Solution

The checkout API endpoint (`POST /checkouts/confirm`) gains an optional `client_email` field. When provided, the back-end creates a client account with a random password and associates the client to the cart after the checkout completes. The `ConfirmRequest` also gains an optional `client_name` field that the storefront can supply; if absent, a name derived from the email prefix is used. The `ConfirmResponse` exposes the cart ID (hidden from JSON output) so downstream post-confirm steps can act on the finished cart.

Client creation is best-effort — it runs after the checkout response is sent and does not fail the checkout if it fails.

## User Stories

1. As an unauthenticated buyer, I want the storefront to optionally send my email during checkout, so that an account is created for me without interrupting the purchase flow.
2. As a returning buyer, I want my purchases tied to my existing email so that the shop can recognize me across sessions.
3. As a shop operator, I want clients auto-created from checkout to appear in the client listing, so I can view their purchase history.
4. As a storefront developer, I want to optionally include a `client_name` during checkout, so the welcome email can address the buyer by name.
5. As a storefront developer, I want the checkout to succeed regardless of whether the client creation step fails, so that purchases are never blocked by account-creation issues.
6. As an operations engineer, I want client creation errors logged clearly so I can diagnose failures without tracing individual checkouts.

## Implementation Decisions

**API contract changes**

The `ConfirmRequest` struct gains two optional fields:
- `client_email` — string, optional. If present and valid, triggers client account creation.
- `client_name` — string, optional. Used as the client's display name. If absent when `client_email` is present, derived from the email prefix.

The `ConfirmResponse` struct gains `CartID` (integer, tagged `json:"-"`) so the handler can act on the finished cart after confirming checkout.

**Client model refactor**

`CreateForCart` is deleted. Its responsibilities are split:

- A new `Create` function creates a user and client record in a transaction. It accepts a pre-computed password hash (bcrypt is heavy — the hash is computed outside any transaction). If the email already belongs to a client, it returns the existing client ID. If the email belongs to a non-client user, it returns an error.
- A new `CreateAndAssociate` function wraps `Create` plus a separate step that sets the cart's `client_id`. It opens its own transaction, independent of the checkout transaction.
- A new query `UPDATE carts SET client_id = $1 WHERE id = $2` (no FINISHED prefix) handles the association without touching `session_id`.

**Password generation**

A new utility function generates a cryptographically random alphanumeric password. It uses `crypto/rand`, not `math/rand`. The plaintext password is generated outside any database transaction and held only long enough for the email step.

**Controller orchestration**

The checkout controller's flow after `Confirm` returns successfully:
1. Send HTTP 200 to the caller immediately.
2. If `client_email` was in the request: generate password, hash it, call `CreateAndAssociate` (best-effort — log errors, do not fail the response).
3. Delegate to the mail module to send the welcome email (best-effort).

**Fallback name**

When `client_name` is empty: extract the portion before `@` in the email address, strip non-alphanumeric characters, and prefix with `usuario-`. Example: `joao.silva@example.com` → `usuario-joaosilva`.

## Testing Decisions

Tests follow the existing integration-test pattern: spin up a test database, hit the API endpoint, and assert database state and HTTP response.

- **Happy path**: Checkout with `client_email` → 200, client row exists, cart's `client_id` is set.
- **Existing client**: Same email on second checkout → 200, same client reused, no duplicate user.
- **Missing email**: Checkout without `client_email` → 200, no client created, `client_id` stays NULL.
- **Invalid email**: `client_email` set to `"not-an-email"` → 400 validation error.
- **Name present**: `client_name` provided → client record uses the provided name.
- **Name absent**: Only `client_email` → client name is the derived fallback.

Test file follows the pattern of `api_cart_checkout_test.go`.

## Out of Scope

- The welcome email itself (covered by PRD "Welcome email via go-mail on client account creation").
- Any modification to the public `/produtos` checkout flow — this is strictly the API checkout.
- Password reset flow or password-change endpoints.
- Associating clients with checkouts retroactively (only new checkouts).
- Phone number capture during checkout.

## Further Notes

This PRD is the prerequisite for the email PRD. The `CreateAndAssociate` function is designed to be called by the controller after `Confirm` returns, and the mail module plugs in at that same point.



Triage: ready-for-agent



## Summary of Changes

### New files
- **util/password.go**: `GenerateRandomPassword` — cryptographically random alphanumeric password using `crypto/rand`.
- **util/name.go**: `DeriveClientName` — derives a display name from an email prefix (e.g. `joao.silva@example.com` → `usuario-joaosilva`).
- **api_checkout_client_test.go**: Integration tests covering happy path, existing client reuse, missing email, custom name, and best-effort failure resilience.

### Modified files

- **query/client.go**: Added `UpdateCartClientIDQuery` — updates only `client_id` on a cart without touching `session_id`.

- **models/clientmodel/model.go**: 
  - Replaced `CreateForCart` with two new functions:
    - `Create`: Creates user + client in a transaction. Returns existing client ID if email already registered as a client. Errors if email belongs to a non-client user.
    - `CreateAndAssociate`: Wraps `Create` + `SetCartClientID` in its own transaction.
  - Added `SetCartClientID` helper for post-checkout cart association.
  - Changed `Phone` field to `*string` to handle NULL telephone (clients created during checkout have no phone).

- **models/clientmodel/resources.go**: Updated `ToResource` to handle nullable phone pointer.

- **models/checkoutmodel/checkout.go**: 
  - Added optional `ClientEmail` and `ClientName` fields to `ConfirmRequest`.
  - Added `CartID` (json:"-") to `ConfirmResponse`.
  - Propagated `CartID` through the `Confirm` function.

- **controllers/checkout/checkout.go**: 
  - After sending HTTP 200 response, calls `createClientAfterCheckout` if `client_email` is present.
  - Generates random password, hashes it (bcrypt), calls `CreateAndAssociate` best-effort — logs errors, never fails the checkout.

### API contract
- **POST /checkouts/confirm** now accepts optional `client_email` and `client_name` in the JSON body.
- Client account creation is best-effort and non-blocking.
- No breaking changes to existing endpoints.
