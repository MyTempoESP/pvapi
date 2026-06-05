---
# pvapi-0bi4
title: Replace external_reference idempotency with content-hash dedup in checkout
status: completed
type: feature
priority: normal
created_at: 2026-06-05T04:13:37Z
updated_at: 2026-06-05T04:23:17Z
---

## Problem Statement

POST /checkouts/confirm uses `external_reference` as the idempotency key — a required field that conflates two concerns: it's both a client-facing order tag AND the system's dedup mechanism. Every caller must invent and send a unique string per checkout. Callers who don't have one (guest checkout, simple integrations) are forced to generate IDs themselves. The dual role also means reusing the same external_reference with different payloads silently returns stale data.

## Solution

Decouple the idempotency mechanism from `external_reference`. Idempotency is now content-driven: a SHA-256 hash of the canonical checkout payload (value_paid + items + client_email). `external_reference` becomes an optional, purely informational tag. The same payload submitted twice returns the same checkout response — regardless of what `external_reference` says.

## User Stories

1. As an API integrator without an order-ID system, I want to POST /checkouts/confirm without providing an external_reference, so that guest purchases don't require me to invent IDs.
2. As an API integrator, I want to safely retry a checkout request after a network timeout, so that I don't accidentally charge the customer twice or double-reserve stock.
3. As an API integrator who uses external_reference as a business tag, I want to be able to reuse the same tag across different purchases, so that my tagging convention isn't blocked by system-level uniqueness constraints.
4. As an API integrator with an existing integration, I want my current requests (which include external_reference) to continue working unchanged, so that the migration is non-breaking.
5. As a support operator, I want every checkout to be traceable by its content fingerprint, so that I can quickly confirm whether two checkouts represent the same real-world purchase.
6. As a QA engineer, I want deterministic idempotency that doesn't depend on client-generated keys, so that test fixtures produce predictable results.

## Implementation Decisions

### Idempotency key: content hash

- SHA-256 over a canonical JSON payload consisting of `{"value_paid":"...", "items":[...], "client_email":"..."}`.
- Items are sorted by `package_id` ascending before marshaling, so item order in the request is irrelevant.
- `client_email` is omitted from the canonical JSON when empty or absent. Absent and `""` produce the same hash.
- `client_name` is NOT included in the hash — it's cosmetic metadata, not transactional identity.

### Schema change

- New column `content_hash TEXT UNIQUE NOT NULL` on the `checkouts` table.
- No backfill needed — production has no checkout data.
- New migration adds the column and the unique constraint.

### Idempotency flow

- INSERT with `ON CONFLICT (content_hash) DO NOTHING RETURNING id, unique_id` inside the existing transaction.
- If RETURNING yields zero rows, fallback SELECT by `content_hash` to fetch the existing checkout.
- Response is identical whether the checkout was freshly created or already existed: `{"success": true, "unique_id": "...", "status": "paid", "qr_code": "..."}`.
- HTTP status remains 200 in both cases.

### external_reference changes

- `binding:"required"` removed from `ConfirmRequest.ExternalReference`.
- Remains stored in the `checkouts.external_reference` column when provided.
- No uniqueness enforcement on external_reference — clients can reuse it across different checkouts.
- Description falls back to `"Checkout — N items"` when external_reference is absent or empty.

### Modules

- **util/checkouthash** (new): canonical JSON marshaling + SHA-256 function. Pure function, testable in isolation.
- **query/checkout.go**: updated `InsertCheckoutQuery` to include content_hash; new `FindCheckoutByContentHashQuery`.
- **models/checkoutmodel/checkout.go**: updated ConfirmRequest, new hashing call, rewritten checkIdempotency, updated writeCheckout, description fallback.
- **controllers/checkout/checkout.go**: no structural changes (model handles the new logic).
- **Migration**: new up/down migration for `content_hash` column + UNIQUE constraint.

## Testing Decisions

### What makes a good test

- Only test external behavior: what the API returns for given inputs, not what hash value was computed.
- Follow existing patterns in `api_checkout_client_test.go` and `api_checkout_shop_test.go` — use the harness, create fixtures, POST confirm, assert response.

### Modules to test

- **util/checkouthash**: unit tests for canonical JSON generation — sorting, field omission, deterministic output.
- **API-level idempotency**: identical payload POSTed twice → same unique_id, same qr_code, status paid on both responses.
- **API-level non-idempotency**: different value_paid or different items → different unique_id.
- **Optional external_reference**: POST without external_reference → 200, description contains "Checkout". POST with empty string → same behavior as absent.
- **Duplicate external_reference different payloads**: POST same external_reference with different items → two distinct checkouts, both succeed.

### Prior art

- `api_checkout_client_test.go` for checkout confirm tests with client email
- `api_checkout_shop_test.go` for checkout confirm with financial breakdown
- `util/util_test.go` for isolated utility function tests

## Out of Scope

- Removing `external_reference` from the database schema
- Changing the structure of `ConfirmItem`
- Adding a standard `Idempotency-Key` header — content-hash is the sole dedup mechanism for now
- Retroactively hashing historical checkout data (none exists in production)
- Client SDK changes

Triage: ready-for-agent



## Summary of Changes

### Files added
- **util/checkouthash/checkouthash.go** — SHA-256 content hash function over canonical JSON (value_paid + sorted items + optional client_email)
- **util/checkouthash/checkouthash_test.go** — unit tests for deterministic hashing, item sorting, email omission
- **api_checkout_dedup_test.go** — API-level tests: identical payload idempotency, different value/items produce distinct checkouts, same external_reference with different payloads, missing/empty external_reference
- **internal/database/migrations/000042_add_checkout_content_hash.up.sql** — adds content_hash TEXT UNIQUE NOT NULL column; drops external_reference uniqueness
- **internal/database/migrations/000042_add_checkout_content_hash.down.sql** — reverts the migration

### Files modified
- **models/checkoutmodel/checkout.go** — replaced external_reference idempotency with content-hash dedup (INSERT with ON CONFLICT DO NOTHING + fallback SELECT); made external_reference optional; description fallback when empty
- **query/checkout.go** — InsertCheckoutQuery includes content_hash with ON CONFLICT; renamed FindCheckoutByExternalRefQuery to FindCheckoutByContentHashQuery

### Verified
All 6 API-level integration tests and all utility unit tests pass.
