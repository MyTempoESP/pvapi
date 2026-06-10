---
# pvapi-e9q3
title: API Token Authentication for Checkout
status: completed
type: feature
priority: normal
created_at: 2026-06-10T21:18:20Z
updated_at: 2026-06-10T21:36:37Z
---

## Problem Statement

Integration partners (external systems, POS terminals, kiosks) need to create checkouts via the PvAuto API. Today, `POST /checkouts/confirm` is fully public — any anonymous caller can reserve stock and generate QR codes. There is no way to identify which integration partner created a checkout, no way to revoke access to a specific partner, and no audit trail linking checkouts to their origin.

## Solution

Add a simple token-based authentication layer for the checkout creation endpoint. Integration partners register as API clients and receive a bearer token. This token must be presented on every checkout creation request. The token identity is frozen into the checkout for audit. Admin users (JWT-authenticated) manage API clients through a CRUD interface.

## User Stories

1. As an integration partner, I want to authenticate with a bearer token, so that I can create checkouts programmatically.
2. As an integration partner, I want my token identity recorded on every checkout I create, so that there is an audit trail linking checkouts to my integration.
3. As an integration partner, I want a clear 401 response when my token is missing or invalid, so that I can diagnose auth failures quickly.
4. As an admin, I want to create API clients with a name and email, so that I can provision new integration partners.
5. As an admin, I want to receive the plaintext token exactly once upon creation, so that I can securely deliver it to the integration partner.
6. As an admin, I want to list all API clients and their status (active/revoked), so that I can audit who has integration access.
7. As an admin, I want to view a single API client's details, so that I can verify when it was created.
8. As an admin, I want to revoke an API client's token, so that I can cut off access for a compromised or decommissioned integration.
9. As an admin, I want to see which API client created a given checkout, so that I can trace issues back to the source integration.
10. As the platform, I want token validation to be fast (hash lookup, not password hashing), so that checkout latency is not degraded.
11. As the platform, I want tokens stored as SHA-256 hashes, so that a database leak does not expose working tokens.
12. As the platform, I want token revocation to be immediate (checked on every request), so that there is no window where a revoked token still works.
13. As the platform, I want anonymous checkout creation to be fully blocked once tokens are deployed, so that every checkout is attributable.

## Implementation Decisions

### Authentication model

API clients are standalone entities — not shops, not users. An API client has a name, email, and a single token. The token is a random 256-bit value with a "pv_" prefix, hex-encoded (64 characters + prefix). The middleware differentiates API tokens from JWT tokens by the prefix: JWT tokens begin with "eyJ", API tokens with "pv_".

Token validation: the incoming token is SHA-256 hashed and looked up in the `api_clients` table. Only active (non-revoked) tokens are accepted. No expiry — tokens live until explicitly revoked.

### Token format

Tokens generated via `crypto/rand`, 32 bytes → hex-encoded → prefixed with "pv_". Example: `pv_a1b2c3d4e5f6...` (64 hex chars after prefix). The prefix makes tokens immediately distinguishable from JWT in logs and configuration.

### Middleware

A new middleware package provides a single function: `Require(pool) gin.HandlerFunc`. It:
1. Extracts the `Authorization: Bearer <token>` header
2. Rejects missing or malformed headers with 401
3. Distinguishes API tokens ("pv_" prefix) from JWT tokens ("eyJ" prefix) — only API tokens are accepted on checkout routes
4. Hashes the token with SHA-256, looks it up in the DB
5. Rejects invalid or revoked tokens with 401
6. Injects `api_client_id` into the gin context on success

The checkout route uses only this middleware — no JWT fallback. Dashboard users create checkouts through the shop interface, not this endpoint.

### Data model

New table `api_clients`:
- `id SERIAL PRIMARY KEY`
- `name TEXT NOT NULL`
- `email TEXT NOT NULL`
- `token_hash TEXT NOT NULL UNIQUE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `revoked_at TIMESTAMPTZ` (nullable; non-null = revoked)

`checkouts` table gains:
- `api_client_id INTEGER REFERENCES api_clients(id) ON DELETE SET NULL` (nullable)

Nullable because historical checkouts from the public era have no API client. New checkouts via the token-gated endpoint always set it.

### Checkout model changes

`checkoutmodel.Confirm()` gains a new parameter `apiClientID int` that is written to the `checkouts.api_client_id` column. The existing `ConfirmRequest` struct and all other checkout request/response fields are unchanged.

### Admin endpoints (JWT-protected, under `/admin`)

All endpoints inherit JWT authentication from the existing admin route group:

- `POST /admin/api-clients` — Create an API client. Request: `{ name, email }`. Response: `{ success, api_client: { id, name, email }, token: "pv_..." }`. The token is returned once in plaintext.
- `GET /admin/api-clients` — List all API clients with id, name, email, created_at, revoked_at.
- `GET /admin/api-clients/:id` — View a single API client.
- `DELETE /admin/api-clients/:id` — Revoke (sets `revoked_at` to now). Idempotent.

### Route changes

- `/checkouts/confirm` — moved from public to token-gated. Applies `apitoken.Require` middleware.
- `/admin/api-clients` — new CRUD group under the existing JWT-protected `/admin` group.

### Database migration

Single migration (000045): creates `api_clients` table, adds `checkouts.api_client_id` column with FK constraint. Down migration reverts both.

### Modules

- `models/apitokenmodel/` — Deep module. Token generation, hashing, CRUD operations over the `api_clients` table. No HTTP dependency.
- `middleware/apitoken/` — Deep module. Single `Require(pool) gin.HandlerFunc`. Token extraction, validation, context injection.
- `controllers/apitoken/` — Thin HTTP glue for admin CRUD handlers.
- `models/checkoutmodel/` — Modified: `Confirm()` accepts `apiClientID int`.
- `controllers/checkout/` — Modified: handler reads API client from context, passes to model.
- `query/` — Modified: SQL constants for `api_clients` CRUD; `InsertCheckoutQuery` accepts `api_client_id`.
- `routes.go` — Modified: new admin sub-routes, middleware applied to checkout route.

### Content hash and dedup

No changes. The content hash mechanism and idempotency behavior are preserved as-is.

## Testing Decisions

Tests follow the codebase conventions: pure unit tests for logic with no external dependencies, and integration tests using the `apiHarness` (real router, real DB pool) for end-to-end HTTP behavior.

### What makes a good test

Tests exercise external behavior (HTTP status codes, response bodies, context values) and model-level contracts (token format, hash determinism). Implementation details like internal helper functions are not tested directly — they are covered by testing their public callers.

### New unit tests

- `models/apitokenmodel/` — Token generation produces "pv_" + 64 hex characters. Token hashing is deterministic. Round-trip: generate → hash → lookup hash matches. Revoked token is rejected.

### New integration tests (added to existing checkout test files)

- Valid token: checkout succeeds, `api_client_id` is written to the checkout.
- Missing `Authorization` header: 401.
- Invalid token (garbage): 401.
- Revoked token: 401.
- JWT token (from login) presented to checkout endpoint: 401 (wrong token type).
- Admin CRUD: create → returns token, list includes new client, revoke → token stops working.

### Updated integration tests

All existing tests that hit `POST /checkouts/confirm` are updated to create an API client fixture (via admin endpoint) and include the bearer token in the request header. Affected test files:
- `api_checkout_dedup_test.go` (10+ subtests)
- `api_checkout_shop_test.go`
- `api_checkout_client_test.go`
- `api_client_test.go`
- `api_client_overview_test.go`
- `api_client_sales_test.go`
- `api_fixtures_test.go` (`generateEarnings`)
- `api_saque_test.go` (if it creates checkouts)

The test harness (`api_harness_test.go`) gains a `createAPIClient` fixture method and `doRequest` is extended to set the `Authorization` header when a token is provided.

## Out of Scope

- Token expiry (tokens live until revoked).
- Multiple tokens per API client.
- Token rotation.
- Scope/rate limiting.
- API client ↔ shop/user binding.
- Non-bearer token transmission (query params, cookies).
- Content hash / dedup changes.
- Dashboard UI for API client management.

## Further Notes

- The "pv_" prefix convention makes tokens self-describing in logs and config. If a token appears in a log entry, it's immediately identifiable as an API integration token rather than a JWT session token.
- Token values should never be logged. The middleware logs only the prefix on validation failure.
- The `ON DELETE SET NULL` on the checkout FK ensures that deleting an API client does not cascade-delete checkout history.
- The admin CRUD endpoints inherit the existing JWT auth on `/admin/*` — no new auth mechanism needed for management.



Triage: ready-for-agent


## Summary of Changes

### Test fixes in api_apitoken_test.go
- Added missing `apitokenmodel` import
- Removed duplicate `asInt` (conflict with api_assert_test.go)
- Removed unused `decodeJSONBody` helper
- Fixed expected status in valid_token_succeeds: 422 → 400 (empty items fails gin validation, but auth passes)

### Query fix in query/apiclient.go
- Added `token_hash` column to FindAPIClientByTokenHashQuery SELECT list. pgx.RowToStructByName requires all struct fields present in result set — missing column caused ValidateToken to always return "token not found".
