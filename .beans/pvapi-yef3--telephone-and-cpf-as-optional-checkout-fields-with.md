---
# pvapi-yef3
title: Telephone and CPF as optional checkout fields with identity resolution
status: todo
type: feature
priority: normal
created_at: 2026-06-05T05:18:22Z
updated_at: 2026-06-05T05:18:31Z
---

## Problem Statement

Clients completing a checkout have no way to provide their telephone or CPF during purchase. Currently only email and name are captured. Without these identifiers, accounts created during checkout lack a telephone for login (relying solely on random passwords) and cannot be linked to a national document (CPF), which limits identity resolution and future account recovery flows.

## Solution

The checkout endpoint gains two new optional fields: telephone and CPF. When provided:

- **Telephone** normalizes to digits and becomes the client's password (for new accounts only; existing passwords are never overwritten). The telephone is always stored/updated on the client record.
- **CPF** (when present and valid as 11 digits) provides identity. If no email is provided, CPF generates a synthetic email (`<cpf>@br`) that serves as the account identifier. No email is sent to synthetic addresses.
- **Telephone-only** (no email, no CPF) generates a synthetic email (`<phone>@tele`) so an account can still be created. The password is the phone digits.
- **Both CPF and telephone, no email**: CPF drives identity (`<cpf>@br`), telephone drives password.
- **Real email always wins**: When a real email is present alongside telephone or CPF, email is the account identity and triggers the normal confirmation email flow. Telephone (if present) still becomes the password for new accounts. CPF is stored as metadata.

A new resolver module centralizes the normalization, validation, and identity-resolution logic in a pure function, making it independently testable.

## User Stories

1. As a client completing a checkout without an account, I want to provide my telephone so that it becomes my password for easier login.
2. As a client completing a checkout without an account, I want to provide my CPF so that my purchases are linked to my national document identity.
3. As a client providing only my CPF (no email), I want an account created with a synthetic email derived from my CPF, so that I have an identity even without a real email address.
4. As a client providing only my telephone (no email, no CPF), I want an account created with a synthetic email derived from my phone number, so that I can log in later using my phone number as both identifier and password.
5. As a client providing both a real email and my telephone, I want my telephone to become my password (instead of a random one) and to receive my confirmation email normally.
6. As a client providing both my real email and my CPF, I want my CPF stored on my account while email drives communication and identity as usual.
7. As a returning client, I want my telephone and CPF updated on my existing account when I provide them in a new checkout, so my contact data stays current.
8. As a returning client, I do NOT want my existing password overwritten when I provide a telephone in a subsequent checkout.
9. As a client whose CPF generates a synthetic email, I do NOT want the system to attempt sending an email to `<cpf>@br` since it is not a real address.
10. As a client whose telephone generates a synthetic email, I do NOT want the system to attempt sending an email to `<phone>@tele`.
11. As a shop operator, I want telephone and CPF stored on client records so I can contact clients and verify identity when needed.
12. As a client whose CPF is formatted incorrectly (not 11 digits), I want the checkout to still succeed — the CPF is silently ignored and the system falls through to the next available identity source (telephone, or no account).
13. As a client whose telephone is all non-digit characters after normalization, I want it treated as not-provided so the checkout proceeds normally.
14. As a client providing both CPF and a custom name, I want my provided name used for the account rather than a derived fallback.
15. As a storefront developer, I want telephone and CPF to be optional in the checkout request so that existing integrations continue working unchanged.
16. As a developer, I want telephone and CPF excluded from the content hash used for checkout deduplication so that the same purchase with different contact details is recognized as a duplicate.
17. As an operator concerned with data privacy, I do NOT want CPF or telephone values logged, since they are sensitive PII.
18. As a returning client who provided CPF in a past checkout, I do NOT want my CPF overwritten with empty when I make a new checkout without providing CPF.

## Implementation Decisions

### Schema change

A new `cpf TEXT DEFAULT NULL` column is added to the `clients` table via migration. The existing `telephone` column already exists and is reused.

### API contract

The checkout confirm request gains two optional string fields: `telephone` and `cpf`. Both are optional — absent or empty means not provided. No validation binding tags. The response is unchanged (no PII echoed back).

### Identity resolver (deep module)

A new pure function resolves all checkout-client identity decisions:

- **Input**: `(email, telephone, cpf, name)` — raw strings from the request
- **Output**: effective email, plaintext password, normalized telephone, validated CPF (empty if invalid), client name, and a boolean `isSyntheticEmail`
- **Normalization**: telephone → digits-only via `[^0-9]` deletion; empty if result is blank
- **CPF validation**: after stripping non-digits, must be exactly 11 digits; otherwise treated as not-provided
- **Password**: if normalized telephone is non-empty → password = phone digits; otherwise → cryptographically random alphanumeric password
- **Effective email**: real email if provided → that email; CPF if valid and no email → `<cpf>@br`; telephone if valid and no email/CPF → `<phone>@tele`; otherwise empty (no account)
- **Synthetic flag**: true when the effective email was derived from CPF or telephone (not a real email provided by the client)
- **Name**: `client_name` if provided → that name; otherwise derived from effective email via `DeriveClientName` (which naturally produces `usuario-<cpf>` or `usuario-<phone>` from synthetic emails)

This module has no database or HTTP dependencies and is tested in isolation with table-driven tests.

### Checkout model changes

`ConfirmRequest` gains `Telephone` and `CPF` string fields. The effective email synthesized by the resolver is set on `ClientEmail` before `Confirm()` is called, so it participates in content-hash deduplication.

### Controller flow

1. Controller calls the resolver with raw request fields, obtaining the `CheckoutClientOutput`
2. Sets `req.ClientEmail = output.EffectiveEmail` (may be synthetic)
3. Calls `checkoutmodel.Confirm()` as before
4. On success, sends HTTP 200 response
5. Only if `EffectiveEmail` is non-empty, calls `sendConfirmationAfterCheckout` with the resolver output
6. `sendConfirmationAfterCheckout` receives `CheckoutClientOutput` instead of the raw request — it no longer generates passwords or derives names internally
7. The mailer is skipped when `isSyntheticEmail` is true
8. If telephone/CPF were provided on an existing client, the UPDATE runs inside `CreateAndAssociateByCheckout`

### Client model changes

`Create` signature gains `telephone, cpf string` parameters. `CreateClientQuery` is updated to INSERT both `telephone` and `cpf`, with `ON CONFLICT DO UPDATE` setting both columns on conflict.

`CreateAndAssociateByCheckout` gains telephone/cpf parameters. After `Create` returns, if `isNew` is false and telephone or CPF were provided (non-empty after normalization), an explicit UPDATE sets only the provided fields on the existing client row — existing data is never overwritten with empty values.

`DeriveClientName` requires no signature changes — it already operates on the effective email string.

### Content hash

Telephone and CPF are excluded from the content hash. The effective email (which may be synthetic) participates normally via the existing `ClientEmail` path.

### Logging policy

CPF and telephone values are never included in log statements. Errors are traceable by `checkout_id` alone.

### CPF in client listing

CPF is excluded from the `/clientes` API listing response. It is stored but not exposed via this endpoint. This may be revisited in collaboration with the front-end team.

## Testing Decisions

Tests follow the existing integration-test pattern from the checkout client test suite: spin up a test database via the API harness, hit the checkout endpoint, and assert database state and HTTP response.

**Resolver unit tests** (new, standalone file): table-driven tests covering normalization, CPF validation, synthetic email generation, password selection (phone vs random), name fallback, and the complete decision matrix combining email/telephone/CPF/name inputs.

**Integration tests** (new file or extension of existing checkout client tests):

- **Telephone as password**: checkout with email + telephone → account created, password = phone digits, confirmation email sent with phone as password
- **CPF-only**: checkout with CPF, no email → account created with `<cpf>@br`, no email sent, name = `usuario-<cpf>`
- **Telephone-only**: checkout with telephone, no email → account created with `<phone>@tele`, password = phone digits, no email sent
- **CPF + telephone, no email**: account identity = `<cpf>@br`, password = phone digits, both stored on client, no email sent
- **Real email wins**: checkout with email + CPF + telephone → email used for identity, email sent, phone = password, CPF stored
- **Existing client — phone update**: second checkout with same email but a new telephone → existing account reused, password unchanged, telephone updated on client
- **Existing client — CPF update**: second checkout after CPF was already set, providing no CPF → existing CPF preserved, not overwritten
- **Existing client — provided fields only**: new checkout provides CPF but no phone → only CPF updated, phone left as-is
- **Invalid CPF fallback**: CPF is non-numeric + telephone → CPF ignored, account uses `<phone>@tele`, checkout still succeeds
- **Empty telephone after strip**: telephone contains only non-digit characters → treated as not-provided
- **No PII in response**: response body does not contain telephone or CPF
- **Dedup**: two identical checkouts with same items but different telephone → second call returns the first checkout's `unique_id`

## Out of Scope

- Front-end changes (storefront form fields for telephone and CPF)
- CPF exposure in the `/clientes` listing endpoint (deferred for front-end discussion)
- Password reset flow using telephone
- Account recovery using CPF
- CPF check-digit validation (only format validation: 11 digits)
- Any modification to the public `/produtos` checkout flow — API checkout only
- Telephone or CPF validation as 400 errors that block checkout completion

## Further Notes

This feature builds on the client auto-creation infrastructure established in the previous checkout client PRD. The resolver is designed as a self-contained pure function to enable isolated testing and to keep the controller thin.

Triage: ready-for-agent
