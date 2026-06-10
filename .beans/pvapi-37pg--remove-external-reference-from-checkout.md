---
# pvapi-37pg
title: Remove external_reference from checkout
status: completed
type: task
priority: normal
created_at: 2026-06-10T23:03:52Z
updated_at: 2026-06-10T23:16:27Z
---

## Problem Statement

The `external_reference` field on checkout confirm conflated two concerns: it was both a client-facing order tag and the system's idempotency key. After pvapi-0bi4, idempotency was decoupled to content-hash dedup, making `external_reference` purely informational and optional. It now serves no purpose — callers have no reason to provide it, and the system gains nothing from storing it. The field adds noise to the API surface and to database queries.

## Solution

Remove `external_reference` entirely from the checkout system: the API contract, the Go model, the SQL queries, and the database schema. The checkout description becomes a static label. Test infrastructure gains a shared `confirmCheckout` helper to reduce boilerplate.

## User Stories

1. As an API integrator, I want the checkout confirmation endpoint to accept only the essential fields (value_paid + items + optional client identification), so that my integration code is simpler and I never need to invent meaningless IDs.
2. As an API integrator with an existing integration that still sends `external_reference`, I want the field to be silently ignored rather than rejected, so that my integration doesn't break after the change.
3. As a developer working on checkout test coverage, I want a shared test helper for checkout confirmation so that I don't duplicate payload construction across test files.

## Implementation Decisions

### API contract change

- `external_reference` is removed from the `POST /checkouts/confirm` request schema. The JSON decoder silently ignores the field if a caller still sends it — no breaking change.
- The checkout description column is set to a static `"Checkout — N items"` label.

### Schema change

- New migration drops the `external_reference` column from the `checkouts` table.
- The migration number is `000046`, following the last existing migration (`000045`).

### Model changes

- `ConfirmRequest.ExternalReference` field removed from the struct.
- `insertCheckoutContentHash` no longer accepts an `externalReference` parameter.
- `buildDescription` simplified to a single format: `"Checkout — N items"`.

### Query changes

- `InsertCheckoutQuery` no longer references the `external_reference` column.

### Test infrastructure

- A new `confirmCheckout` helper function is added to the shared test harness. It accepts `value_paid`, `items`, and optional `extras` (client_email, client_name, telephone, cpf), builds the request body, POSTs to `/checkouts/confirm`, and returns the raw response recorder for the caller to assert.
- Three dedup tests that specifically verify `external_reference` behavior are dropped:
  - `same_external_reference_different_payloads_are_distinct`
  - `missing_external_reference_succeeds`
  - `empty_external_reference_succeeds`
- The shop test's raw SQL lookups that used `chk.external_reference` are updated to use `chk.unique_id` from the checkout response.
- All checkout confirmation calls across test files are refactored to use the new helper, eliminating all `external_reference` references from test code.

## Testing Decisions

### What makes a good test

- Tests assert external behavior (HTTP status, response structure, idempotency) — never internal hashes or database column values.
- Existing checkout tests continue to pass with the same assertions; only the payload construction changes.
- The idempotency tests (identical payload → same unique_id, different payload → different unique_id) retain their semantics unchanged since idempotency is content-hash-driven.

### Modules tested

- All existing checkout API-level integration tests are updated. No new tests are added — this is a removal, not a feature.

### Prior art

- `api_checkout_client_test.go` for checkout confirm patterns with client identification
- `api_checkout_dedup_test.go` for content-hash idempotency verification
- `api_checkout_shop_test.go` for financial breakdown verification
- `api_harness_test.go` for shared test helpers

## Out of Scope

- Changing the checkout description format beyond the static label
- Removing the `description` column from the schema
- Adding an `Idempotency-Key` header
- Changing the content-hash idempotency mechanism
- Modifying old migrations (history is preserved as-is)

## Further Notes

- This is a cleanup following pvapi-0bi4 (content-hash dedup), which already made external_reference optional and removed its uniqueness constraint.
- Production has no checkout data, so no backfill or data migration is needed.

Triage: ready-for-agent



## Summary of Changes

### Database
- Created migration  to drop  column from  table

### Model
- Removed  field from  struct
- Removed  parameter from 
- Simplified  to static format: "Checkout — N items"

### Query
- Removed  from  INSERT/VALUES clauses

### Test Infrastructure
- Added  shared helper to 
- Removed three dedup tests specific to external_reference behavior
- Removed all  from test payloads across all test files
- Updated  SQL queries to use  instead of 
- Refactored several tests to use the new  helper

### Database
- Created migration `000046` to drop `external_reference` column from `checkouts` table

### Model
- Removed `ExternalReference` field from `ConfirmRequest` struct
- Removed `externalReference` parameter from `insertCheckoutContentHash`
- Simplified `buildDescription` to static format: "Checkout — N items"

### Query
- Removed `external_reference` from `InsertCheckoutQuery` INSERT/VALUES clauses

### Test Infrastructure
- Added `confirmCheckout` shared helper to `api_fixtures_test.go`
- Removed three dedup tests specific to external_reference behavior
- Removed all `external_reference` from test payloads across all test files
- Updated `TestCheckoutShopClientID` SQL queries to use `unique_id` instead of `external_reference`
- Refactored several tests to use the new `confirmCheckout` helper
