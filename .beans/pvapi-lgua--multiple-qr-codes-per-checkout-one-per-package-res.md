---
# pvapi-lgua
title: Multiple QR Codes per Checkout (one per package_reserved)
status: completed
type: feature
priority: normal
created_at: 2026-06-05T19:15:31Z
updated_at: 2026-06-05T19:27:38Z
---

## Problem Statement

The checkout endpoint currently returns a single QR code encoding the checkout's `unique_id`. This means a single QR code represents the entire order, but at pickup, different packages may be at different pickup locations or handled by different staff. A single QR code cannot be split across multiple pickup counters.

## Solution

Generate one QR code per `package_reserved` row, each encoding that row's own UUID. Return all QR codes in the checkout response as a `qr_codes` array pairing each `package_id` with its QR code. Remove the old single `qr_code` field.

At the same time, remove the pickup confirmation email from the checkout flow — multiple QR codes make the email cumbersome, and scanning happens externally. Client auto-creation on checkout is preserved.

## User Stories

1. As an integration partner, I want to receive one QR code per reserved package in the checkout response, so that I can distribute individual QR codes to different pickup counters or staff members.
2. As an integration partner, I want each QR code to map directly to a package_id, so that I know which QR code corresponds to which product at which pickup location.
3. As an integration partner, I want the checkout response to no longer include the deprecated `qr_code` field, so that the API contract is clean and unambiguous.
4. As an integration partner, I want the checkout `unique_id` to still be present in the response, so that I can correlate the entire checkpoint transaction.
5. As an integration partner, I want checkout idempotency to still work when the same content hash is submitted again, returning QR codes for all existing package_reserved rows.
6. As a returning customer, I want my account to still be auto-created or associated on checkout, so that my purchase history and identity (CPF/telephone) persist across purchases.
7. As a returning customer, I want confirmation emails to no longer be sent on checkout, so that I am not overwhelmed by emails with multiple QR codes.
8. As a developer, I want the mail package to remain intact in the codebase, so that it can be re-enabled later without rebuilding from scratch.
9. As an API consumer reading the documentation, I want the `llms.txt` and OpenAPI spec to reflect the new `qr_codes` response format and the removal of email delivery.
10. As a QA engineer, I want existing tests updated to assert on `qr_codes` instead of `qr_code`, so that CI stays green.

## Implementation Decisions

### Schema change
- Add `unique_id UUID NOT NULL DEFAULT gen_random_uuid()` to `package_reserved` via a new migration.
- The `package_reserved.status` column remains unchanged (`'pending'` on insert). The external scanning system manages status transitions.

### Query changes
- `InsertPackageReservedQuery`: append `RETURNING unique_id, package_id` to the INSERT clause. The model layer collects these results to build QR codes.
- New query `FindPackageReservedByCheckoutQuery`: `SELECT unique_id, package_id FROM package_reserved WHERE checkout_id = $1`. Used in the dedup path to regenerate QR codes for existing checkouts.

### Response format
- New struct `QRCodeEntry` with `PackageID int` and `QrCode string` (base64 data URI).
- `ConfirmResponse` gains `QrCodes []QRCodeEntry` (JSON: `qr_codes`). The old `QrCode string` field is removed.
- Response JSON shape:
  ```json
  {
    "success": true,
    "message": "checkout confirmed",
    "unique_id": "<checkout uuid>",
    "status": "paid",
    "qr_codes": [
      {"package_id": 42, "qr_code": "data:image/png;base64,..."}
    ]
  }
  ```

### QR code payload
Each QR code encodes the `package_reserved.unique_id` (UUID), same pattern as the checkout QR code today. No composite key, no URL — just the UUID as lookup key.

### Checkout flow (new path)
After inserting `package_reserved` rows with `RETURNING`, collect the `(unique_id, package_id)` pairs. Generate one QR code per UUID via `qrcode.Encode()`. Build the `[]QRCodeEntry` slice and return it.

### Checkout flow (dedup path)
When a duplicate content hash is detected, query `FindPackageReservedByCheckoutQuery` to get all `(unique_id, package_id)` pairs for the existing checkout. Generate QR codes from each UUID and return them.

### Client association (post-checkout)
- The function is renamed to `associateClientAfterCheckout`.
- Email sending and all email-specific data resolution (`ResolvePickupEmailData`, `pickupEmailRow`, `formatPickupAddress`, `env.SMTPLoginLink`, `env.SMTPResetLink`) are removed.
- Client auto-creation via `clientmodel.CreateAndAssociateByCheckout` is preserved.
- The `mail.Mailer` parameter is removed from the checkout handler signature and from `routes.go`.

### Mail package
The entire `mail/` package is kept unchanged — `Mailer` interface, `Client`, template, types, all intact.

### Documentation
- `Dev/llms.txt`: update checkout response section to document `qr_codes` array, remove email delivery documentation, remove references to single `qr_code`.
- `Dev/requests_implemented/api-integrations.yaml`: update `RespostaCheckout` schema to include `qr_codes` as array of `{package_id, qr_code}`, remove `qr_code` field. Update description to remove email delivery mentions.

## Testing Decisions

- Tests should assert on external behavior only: the JSON response shape returned by `POST /checkouts/confirm`.
- No new test files needed. Existing tests are updated in place.
- Dedup tests (`api_checkout_dedup_test.go`): assertions that compared `qr_code` strings now compare `qr_codes` arrays (same entries = same checkout).
- Client creation tests (`api_checkout_client_test.go`): assertions that checked `email.QrCode` are removed; tests now verify no confirmation email is sent.
- The testing approach follows existing patterns: `apiHarness` with `doJSON`, `decodeJSON`, `asSlice`, `asMap`, `asString` helpers.

## Out of Scope

- Changes to how packages are scanned or confirmed at pickup (handled externally).
- Changes to `package_reserved.status` lifecycle beyond initial `'pending'`.
- Re-enabling confirmation emails. The mail package is kept but not wired.
- Any changes to the `/clientes` or `/produtos` endpoints.
- Changes to the idempotency mechanism (content hash logic is unchanged).

## Further Notes

The `client_email` field is still accepted in the request body and still participates in the content hash for dedup, even though no email is sent. This preserves existing dedup behavior.

## Summary of Changes

### Migration
- Added `unique_id UUID NOT NULL DEFAULT gen_random_uuid()` to `package_reserved` (000044)

### Queries
- Added `RETURNING unique_id, package_id` to `InsertPackageReservedQuery`
- Added `FindPackageReservedByCheckoutQuery`

### Model
- Added `QRCodeEntry` struct with `PackageID` and `QrCode` fields
- Changed `ConfirmResponse.QrCode string` → `QrCodes []QRCodeEntry`
- `Confirm()` now generates one QR code per `package_reserved` row using its own `unique_id`
- Dedup path regenerates QR codes from existing `package_reserved` rows
- Removed email-related functions (`ResolvePickupEmailData`, `formatPickupAddress`, `pickupEmailRow`)

### Controller
- Removed `mail.Mailer` parameter from `ConfirmHandler`
- Renamed `sendConfirmationAfterCheckout` → `associateClientAfterCheckout`
- Removed all email sending logic (mail package kept intact per spec)
- Response JSON: `qr_code` → `qr_codes`

### Routes
- Removed `mailer` parameter from `setupRoutes` and `ConfirmHandler` call

### Tests
- Updated all assertions from `qr_code` to `qr_codes`
- Removed email-assertion tests (email no longer sent)
- Removed mailer from test harness

### Documentation
- Updated `Dev/llms.txt` checkout response section
- Updated `Dev/requests_implemented/api-integrations.yaml` schema

Triage: completed
