---
# pvapi-hpmh
title: Rename package_id to id in /checkouts/confirm request items
status: todo
type: task
created_at: 2026-06-18T08:28:52Z
updated_at: 2026-06-18T08:28:52Z
---

## Problem Statement

The `/checkouts/confirm` endpoint requires clients to send `package_id` in each item, exposing an internal database concept (`packages` table). Clients already interact with packages through the products listing endpoint, where each pickup under a product has a `package_id` — but from the client's perspective, this is just an opaque variant ID they select and send back. Exposing the internal "package" naming leaks implementation detail.

## Solution

Rename the JSON field from `package_id` to `id` in `ConfirmItem` (the request DTO for `/checkouts/confirm`). The Go struct field name stays `PackageID` — only the JSON wire format changes.

## User Stories

1. As an API client, I want to send `id` instead of `package_id` in checkout confirm items, so that I don't need to know about internal database table names.
2. As an API client, I want the confirm request to use the same opaque identifier I received from the products listing endpoint, without a naming mismatch between endpoints.

## Implementation Decisions

- **JSON tag only**: Rename `ConfirmItem.PackageID`'s json tag from `"package_id"` to `"id"`. The Go field name stays `PackageID` — it maps to `packages.id` in the DB and the field name accurately describes internal use.
- **Response unchanged**: `QRCodeEntry.PackageID` keeps its `"package_id"` JSON tag. The QR codes array is metadata the client stores, not something they reconcile against the request payload. Consistency with request field name is unnecessary.
- **No DB changes**: All queries (`InsertPackageReservedQuery`, `CheckPackageAvailabilityQuery`) reference `packages.id` by column name, not by JSON field name. Unaffected.
- **No route/gate/controller changes**: The controller passes the deserialized `ConfirmRequest` to `checkoutmodel.Confirm`, which uses `item.PackageID` as a Go int. Unaffected.
- **Test fixtures**: All test JSON payloads that construct `{"package_id": ..., "quantity": ...}` must use `{"id": ..., "quantity": ...}` instead.

## Testing Decisions

- Update all existing integration test fixtures in `api_checkout_client_test.go`, `api_checkout_cancel_test.go`, `api_checkout_shop_test.go`, and `api_checkout_dedup_test.go` to use `"id"` instead of `"package_id"`.
- No new tests needed — the existing test suite fully covers the confirm flow. The rename is a wire-format change that existing tests will validate implicitly.
- Run `test-code` skill targeting checkout tests to verify.

## Out of Scope

- Renaming `ProductPickupResource.PackageID` (products listing response) — clients already receive `package_id` there. Changing both would double the breaking surface for no gain.
- Renaming any Go identifiers, database columns, or internal model fields.
