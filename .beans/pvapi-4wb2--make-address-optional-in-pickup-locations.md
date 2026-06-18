---
# pvapi-4wb2
title: Make address optional in pickup_locations
status: todo
type: feature
priority: normal
tags:
    - ready-for-agent
created_at: 2026-06-18T23:16:14Z
updated_at: 2026-06-18T23:16:14Z
---

## Problem Statement

Shop managers need to create pickup locations without providing a full address. Currently the `AddressRequest` struct enforces `binding:"required"` on all address fields, making the address mandatory when creating or updating a pickup location.

The frontend always sends an address object — even when the manager leaves it blank — with all fields as empty strings (`"street": "", "number": "", ...`). The backend must accept this and treat it as "no address".

## Solution

Relax address validation for pickup locations: accept an address payload where all fields are empty strings (meaning "no address") or all fields are filled (normal address). Reject partially filled addresses with a 400 error.

When the address is blank, store a sentinel address row in the database (number=0, all strings empty) and omit the `address` field from the API response for that pickup location.

## User Stories

1. As a shop manager, I want to create a pickup location without providing an address, so that I can register temporary or informal pickup points quickly.
2. As a shop manager, I want to update an existing pickup location to remove its address, so that I can correct or simplify location data.
3. As a shop manager, I want the API to reject a pickup location with a partially filled address (e.g., street provided but number missing), so that I don't accidentally create incomplete records.
4. As an API consumer (consignee, checkout flow), I want pickup locations without an address to still appear in listings and be selectable, so that blank-address pickups remain usable.
5. As a frontend developer, I want the API response to omit the `address` key entirely when a pickup location has no real address, so that the UI can easily distinguish between "no address" and "empty address".

## Implementation Decisions

### Address sentinel

When all address fields are empty strings during create, instead of setting `address_id` to NULL, insert an address row with sentinel values: `number = 0`, all text fields as empty strings. This avoids schema changes (the `number` column is `integer NOT NULL`) and keeps query JOINs working without NULL-handling branches.

### AddressRequestOptional

A new type `AddressRequestOptional` mirrors `AddressRequest` but without `binding:"required"` tags. It has its own `Prepare()` method that converts to `Address`, mapping an empty string for `number` to `"0"` (the string representation used internally).

### All-or-nothing validation

Validation happens in the request `Prepare()` methods:
- If the address pointer is nil, default to an all-empty sentinel address.
- If the address is present, check: either every field is empty, or every required field (street, number, neighborhood, city, state, CEP) is non-empty.
- Any other combination returns a validation error.

### Optional address in API response

`Address` gains:
- `IsEmpty() bool` — returns true when all fields are zero/empty strings.
- `ToResource() *Address` — returns nil when `IsEmpty()`, otherwise a pointer to the address.

`PickupLocationResource.Address` changes from `Address` (value) to `*Address` (pointer) with `json:"address,omitempty"`. When the address is a sentinel, the JSON omits the `address` key entirely.

### No changes to other consumers

Consignee, shop, client-overview, and checkout resources continue to embed `Address` by value. Their address is always meaningful and required. The new `ToResource()` method is only called from the pickup resource.

### No schema changes

`pickup_locations.address_id` is already nullable. The sentinel approach (inserting a real address row with sentinel values) requires no migration.

### No query changes

`CreatePickupLocationQuery` continues to call `insert_new_address()`. When fields are empty strings, the function receives `''` for text and `0` for integer, producing a valid address row.

## Testing Decisions

Tests are integration-style API tests following the existing pattern in `api_pickup_test.go`. Good tests exercise the external HTTP contract: request shape, response shape, status codes.

Test scenarios:
- Create pickup with all-empty address → 200, response omits address field
- Create pickup with partial address (only street filled) → 400
- Create pickup with full address → 200, response includes address
- Update pickup: set address to all-empty → 200, subsequent GET omits address
- Update pickup: set address to full → 200, subsequent GET includes address

## Out of Scope

- Changing address validation for shops, consignees, or any other entity. Only pickup locations are affected.
- Database migration to make address columns nullable.
- UI changes in the frontend.
- Affecting the checkout `QRCodeEntry` or consignment `PickupLocationSummary` — those continue to return address data as-is (sentinel address with number=0 will appear in those responses).

## Further Notes

The `address.Address.Number` field is stored as a `string` in Go but maps to `INTEGER` in PostgreSQL. This is pre-existing behavior — the codebase converts the string at the query parameter level. The sentinel value `"0"` is chosen because 0 is an invalid real-world street number, making it distinguishable if needed.
