---
# pvapi-bqrf
title: PUT /pontos/:pickup — Update pickup location
status: completed
type: feature
priority: normal
created_at: 2026-05-27T15:24:51Z
updated_at: 2026-05-27T15:33:27Z
---

## Problem Statement

Shops can create and list pickup locations (pontos de entrega) via POST and GET on /pontos, but they cannot update a pickup location's name, date/time, or address after creation. Any mistake at creation time forces the shop to create a new pickup and abandon the old one.

## Solution

Add a PUT /pontos/:pickup endpoint, JWT-authenticated and gated for shops. The endpoint accepts a JSON body with the same shape as the POST create payload, allowing shops to update name, datetime, and address of a pickup location they own.

## User Stories

1. As a shop, I want to update the name of my pickup location, so that I can correct a typo or rebrand the point.
2. As a shop, I want to change the date and time of a pickup location, so that I can reschedule logistics.
3. As a shop, I want to update the address fields (street, number, neighborhood, city, state, CEP, complement) of my pickup location, so that I can correct or relocate a pickup point.
4. As a shop, I want to update only some fields (e.g., just the name) without needing to resend the full address, so that partial edits are fast and safe.
5. As a shop, I want the update to fail with 403 if I try to modify a pickup location that belongs to another shop, so that data integrity is preserved.
6. As a shop, I want the update to fail with 404 if the pickup location ID does not exist, so that I get clear feedback.
7. As a shop, I want the update to fail with 400 if the request body is malformed or missing required validation, so that I can fix my payload.
8. As an admin, I want to update any shop's pickup location, so that I can assist with support operations.
9. As a consignee (monitor), I want to be denied access to this endpoint, since pickup management is a shop/admin concern.
10. As an unauthenticated client, I want to receive a 401 when hitting this endpoint, so that only authorized actors can modify pickups.

## Implementation Decisions

### Route

- Add `pickupLocations.PUT(":pickup", pickups.UpdateHandler(pool))` to the existing `/pontos` group in routes.go.
- The `:pickup` param carries the pickup location integer ID.

### Request payload

New type `UpdatePickupLocationRequest` with optional fields matching the create payload but without binding:"required" tags, enabling partial updates:

- `name` — string, optional
- `datetime` — string (ISO 8601), optional
- `address` — pointer to `*address.AddressRequest`, optional (nil means don't update address)

### Controller

New `UpdateHandler` in controllers/pickups/pickups.go, patterned after bankaccount.UpdateHandler:

- Parse `:pickup` param as int (400 on failure).
- Bind JSON to `UpdatePickupLocationRequest` (400 on failure).
- Ownership is enforced by the gate (see below); the controller relies on gate authorization.
- Call model-layer Update.

### Model layer

- New `UpdatePickupLocationRequest` in models/pickupmodel/requests.go.
- New `Find` function in models/pickupmodel/model.go that fetches a PickupLocation by ID (needed to retrieve the address_id for the address update).
- New `Update` function in models/pickupmodel/model.go, patterned after bankaccountmodel.Update:
  - Takes ctx, base, pickupID, and UpdatePickupLocationRequest.
  - Executes the update query with the subset of non-nil fields.

### Queries

- `FindPickupLocationQuery` in query/pickup.go — SELECT pickup_locations + aliased_addresses_view join by ID.
- `UpdatePickupLocationQuery` in query/pickup.go — UPDATE pickup_locations SET name, pickup_at, ... WHERE id.
- Address update query — either a new `UpdateAddressQuery` in query/address.go, or a composite approach: since the pickup location references an address row via address_id, the update must also UPDATE the addresses table. The simplest path is to UPDATE the existing address row in-place (single query with a subquery or two separate statements in a transaction).

### Gate / Authorization

The gate at gate/pickups.go already has `Modify` → `Owns` → `shopmodel.HasPickup` wired for PUT. The route param `:pickup` is already handled by `Dispatch` via `r.Params.ByName("pickup")`.

**Pre-existing bug to fix**: `ShopHasPickupQuery` in query/shop.go references table `pickups`, but the actual table is `pickup_locations`. Fix the query to join on `pickup_locations` instead:

```sql
SELECT (p.id IS NOT NULL) AS hasPickup
FROM shops s
LEFT JOIN pickup_locations p ON p.shop_id = s.id
WHERE s.id = $1 AND p.id = $2
```



### Integration test

- Add two tests to api_pickup_test.go:
  1. **update_pickup**: Create a pickup, PUT update name + datetime, GET list to verify changes.
  2. **update_pickup_cross_shop**: Shop A creates pickup, Shop B attempts PUT update → expect 403.

## Testing Decisions

- **What makes a good test**: Integration tests that hit the real API server against a temp PostgreSQL database. Only test external behavior (status codes, response shape, data persistence). No mocking of DB or auth.
- **Which modules are tested**: The full stack (route → gate → controller → model → query) is exercised end-to-end.
- **Prior art**: Follow the patterns in api_pickup_test.go (creates shop fixture, logs in, does request). The cross-ownership test mirrors patterns from api_saque_test.go and api_bank_account_test.go.

## Out of Scope

- Batch/bulk update of multiple pickup locations.
- Deleting pickup locations (DELETE is gated but has no handler yet — separate issue).
- Public (unauthenticated) access to pickup locations.
- Updating the consignment-pickup relationship (which consignment uses which pickup point).

## Further Notes

- The address update strategy (in-place UPDATE on the addresses table vs insert-new + repoint FK) should be validated against the existing DB schema. The `insert_new_address` function already exists; an in-place UPDATE is simpler and avoids orphaned address rows.
- The `PickupLocationBlueprint` type is used for creates; the update path may use it for consistency or work with a flatter approach (model.Update taking individual params) — pick whichever keeps the model layer cleanest.
- The pickup location unique-name constraint (`pickup_location_unique_name`) may cause 409 conflicts if two pickups share a name within the same shop. This is existing behavior from create and should be handled gracefully (500 → mapped to an appropriate error).

Triage: ready-for-agent



## Summary of Changes

### Bug fix
- query/shop.go: Fixed ShopHasPickupQuery — table name `pickups` → `pickup_locations` (pre-existing bug blocking the gate from finding pickup locations by ID).

### Queries
- query/pickup.go: Added FindPickupLocationQuery (SELECT with address join by ID) and UpdatePickupLocationQuery (UPDATE with COALESCE for partial updates).
- query/address.go: Added UpdateAddressQuery with subquery to find address_id from pickup_locations, enabling in-place address update.

### Model
- models/pickupmodel/requests.go: Added UpdatePickupLocationRequest with optional `*string` name/datetime fields and `*AddressRequest` for partial updates.
- models/pickupmodel/model.go: Added Find (fetches a PickupLocation by ID) and Update (updates pickup_locations + optionally updates addresses).

### Controller
- controllers/pickups/pickups.go: Added UpdateHandler — parses `:pickup` param, binds JSON, calls model.Update, returns success.

### Route
- routes.go: Added `pickupLocations.PUT(":pickup", pickups.UpdateHandler(pool))` to the /pontos group.

### Tests (7 new subtests in TestPickup)
- update_pickup: Full update (name, datetime, address) + GET list verification.
- update_pickup_partial_name: Only name sent, other fields preserved.
- update_pickup_unauthorized: 401 for unauthenticated.
- update_pickup_cross_shop: 403 when Shop B tries to update Shop A pickup.
- update_pickup_bad_param: 403 for non-integer pickup ID (gate catches first).
- update_pickup_not_found: 403 for non-existent pickup ID (gate catches first).
