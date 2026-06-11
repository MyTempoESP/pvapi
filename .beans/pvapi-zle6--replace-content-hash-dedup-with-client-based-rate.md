---
# pvapi-zle6
title: Replace content-hash dedup with client-based rate limiting
status: completed
type: feature
priority: normal
created_at: 2026-06-11T00:26:28Z
updated_at: 2026-06-11T00:39:44Z
---

## Problem Statement

A client bought the same product one week apart — identical value, items, and email. The content-hash-based deduplication treated the second checkout as a duplicate, silently returning the old checkout's confirmation and QR codes without deducting stock or creating a new transaction. The client saw a success response but no actual purchase occurred.

More broadly, the current content-hash mechanism (SHA-256 of value_paid + items + email) is a timeless idempotency guard. It cannot distinguish between a double-submission 2 seconds apart (legitimate idempotency) and a genuine repurchase 1 week later (should be a new checkout).

## Solution

Replace content-hash deduplication with per-client rate limiting: a client cannot create a new checkout within 60 seconds of their last checkout. Within that window, the API returns HTTP 429 with a Portuguese-friendly message showing remaining seconds. Double-submissions are handled client-side (form disables the submit button).

Additionally, make client creation synchronous (moved from fire-and-forget goroutine into the checkout flow), and require at least one identity field (email, CPF, or telephone) on every checkout — anonymous checkouts are no longer supported.

## User Stories

1. As a client buying the same product I bought last week, I want my purchase to succeed as a new checkout, so that I receive new QR codes and the shop gets paid for my new order.

2. As a client who just completed a checkout, I want to be told to wait if I accidentally submit again within 60 seconds, so that I know my first purchase went through and I don't need to redo it.

3. As a client filling a checkout form, I want my double-click on the submit button to be harmless, so that I'm not confused by seeing the same order twice.

4. As a client, I want to know exactly how many seconds I need to wait before I can make another purchase, so that I can plan my next action.

5. As a shop owner, I want every successful checkout to deduct stock and generate fresh QR codes, so that I don't lose sales to silent deduplication failures.

6. As a shop owner, I want legitimate repurchases to always succeed regardless of how identical they are to past purchases, so that I never miss revenue from returning clients.

7. As an API integration partner, I want the rate-limit error to include a retry_after field in seconds, so that I can display a countdown timer to the end client.

8. As an operator, I want client accounts to be created synchronously during checkout (not asynchronously after responding), so that the client record exists immediately and can be used for rate limiting and attribution.

9. As an operator, I want every checkout to require at least one identity field (email, CPF, or telephone), so that every purchase is attributable to a real client and rate limiting is always enforceable.

10. As an operator, I want anonymous checkouts to be rejected with a clear validation error, so that API consumers know they must provide client identity.

## Implementation Decisions

**Rate-limit mechanism** — Per-client globally, querying the checkouts table. Within the checkout transaction, lock the client row with SELECT FOR UPDATE, then query MAX(processed_at) WHERE client_id = . If the latest processed_at is within 60 seconds of now, return 429. Otherwise, proceed.

**Rate-limit response** — HTTP 429 Too Many Requests with JSON body: success: false, message containing remaining seconds in Portuguese, and retry_after field with integer seconds.

**Window duration** — 60 seconds, hardcoded.

**Double-submissions** — Always return the rate-limit error within the window, even if the payload is identical. The form is expected to disable the submit button client-side after first click.

**Synchronous client creation** — move clientmodel.CreateAndAssociateByCheckout from the controller's fire-and-forget goroutine into the checkout model, called before the checkout transaction. If client creation fails, the checkout fails with 500.

**Identity validation** — at least one of client_email, cpf, or telephone must be provided. Return 400 with message "Informe email, CPF ou telefone para continuar" if none present.

**Schema changes** — Single migration that: drops content_hash column and its UNIQUE constraint, drops external_reference column, adds a composite index on checkouts(client_id, processed_at DESC) for the rate-limit query.

**Concurrency** — SELECT FOR UPDATE on the clients row serializes checkouts for the same client, preventing two simultaneous requests from both passing the rate-limit check.

**Removals** — util/checkouthash package deleted entirely. InsertCheckoutQuery simplified to a plain INSERT without ON CONFLICT. FindCheckoutByContentHashQuery removed.

**client_email field** — no longer participates in any dedup hash. It serves purely as an identity field for client resolution.

## Testing Decisions

Tests should exercise external behavior: API-level integration tests against a real database, matching the pattern used by existing api_checkout_*_test.go files.

**Unit test** — The rate-limit check function (given a client_id, query last checkout time, compute remaining seconds) should be tested in isolation. A thin function: takes a DB connection, returns (remaining int, isBlocked bool, error).

**Integration tests** — Rewrite the dedup test file to cover rate-limit scenarios instead of content-hash scenarios: identical payload 2 seconds apart returns 429 (client blocked), different payload within window returns 429, payload after 60 seconds succeeds, no-identity request returns 400.

**Client tests** — Adjust for synchronous client creation: client exists in DB immediately after checkout response, not eventually. Add tests for checkout failure when client creation fails.

## Out of Scope

- Configurable rate-limit window (env var or admin setting)
- Per-shop or per-API-client rate-limit scoping
- IP-based rate limiting for truly anonymous clients
- Client-side form behavior (disabling submit button) — this is frontend work

## Further Notes

The content-hash column and unique constraint were introduced in migration 000042. The external_reference column and its unique constraint were introduced in 000019/000038. The cart_id + external_reference unique constraint from 000021 was already made obsolete when carts were dropped in 000040.



Triage: ready-for-agent



## Summary of Changes

### Schema (migration 000047)
- Dropped  column and its UNIQUE constraint from 
- Added composite index  on  for rate-limit queries

### Removed
- Deleted  package entirely (content-hash dedup is gone)
- Removed  from queries
- Removed  from  (plain INSERT, no ON CONFLICT)
- Removed  goroutine from controller

### New rate-limiting logic
- Added : queries  for a client, returns seconds since last checkout
- Added  (60s, exported for test overrides)
- Added :  on clients row to serialize checkouts
- Added  and  error type (supports  field)
- Rate-limited requests return HTTP 429 with Portuguese message and  seconds

### Synchronous client creation
- Client creation moved from fire-and-forget goroutine into the checkout transaction
- Client is created/resolved before stock validation and checkout insert
- Existing client fields (telephone, CPF) updated inline when provided

### Identity validation
- Controller rejects requests with no , , or  with 400
- Message: "Informe email, CPF ou telefone para continuar"
- Added  error for resolver-level failures (e.g., non-digit-only phone)

### Test changes
- Rewrote  → rate-limit tests (6 subtests: identical/different payload, different client, no-identity, CPF-only, telephone-only)
- Updated : no-email → no-identity test expects 400, removed content-hash dedup test, added rate-limit disabling for tests needing same-client multi-checkout
- Updated  fixture to include identity email
- Updated , , ,  to provide identity or disable rate limiting as needed
