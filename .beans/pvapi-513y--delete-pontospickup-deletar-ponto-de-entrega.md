---
# pvapi-513y
title: DELETE /pontos/:pickup — deletar ponto de entrega
status: completed
type: feature
priority: normal
created_at: 2026-06-12T01:06:51Z
updated_at: 2026-06-12T01:11:20Z
---

## Problem Statement

Shop operators can create, list, and update Pickup Locations (Pontos) via the API, but there is no way to delete a Pickup Location that is no longer needed — e.g., a location that was set up for a one-time event, or an incorrectly registered address. Without a DELETE endpoint, stale Pickup Locations accumulate indefinitely, cluttering the shop's list view and making it harder to manage active locations.

## Solution

Add a `DELETE /pontos/:pickup` endpoint that allows a Shop to permanently remove one of its own Pickup Locations. The operation is a hard delete: the pickup row is removed from the database, and any associations between that pickup and Consignments (through the `consignment_pickup_locations` junction table) are cleaned up automatically via the existing `ON DELETE CASCADE` foreign key.

## User Stories

1. As a Shop operator, I want to delete a Pickup Location I no longer use, so that my list of pickup locations stays clean and accurate.
2. As a Shop operator, I want the deletion to fail if I try to delete another shop's Pickup Location, so that I cannot accidentally or maliciously interfere with other shops.
3. As a Shop operator, I want the deletion to be permanent (hard delete), so that I don't accumulate stale soft-deleted records in the database.
4. As a Shop operator, I want to receive a clear success message after deleting a Pickup Location, so that I know the operation completed.
5. As a Shop operator, I expect any Consignments that had this Pickup Location assigned to simply lose that assignment (the junction-table entry is cascade-deleted), so that existing consignments are not broken or deleted.
6. As an unauthenticated user, I want the endpoint to reject me with 401, so that only authorized shops can delete pickup locations.
7. As a Consignee, I should not be able to delete a Pickup Location, since only the owning Shop manages its own locations.
8. As an Admin, I should not be able to delete an arbitrary shop's Pickup Location directly through this shop-scoped endpoint.

## Implementation Decisions

- **Hard delete**: The `DELETE FROM pickup_locations WHERE id = $1` SQL permanently removes the row. This is consistent with how products, consignments, reports, saques, and bank accounts are deleted in the codebase.
- **No address cleanup**: Each Pickup Location has a dedicated address row (addresses are not shared between pickups, shops, or consignees). Orphaned address rows are harmless.
- **Cascade handled by PostgreSQL**: The `consignment_pickup_locations.pickup_location_id` foreign key has `ON DELETE CASCADE`, so the junction table rows are automatically removed. No application-level cascade logic is needed.
- **Route parameter name**: The URL parameter is named `:pickup` to match the existing PUT `/pontos/:pickup` convention. The authorization gate already extracts this parameter by name, so no gate changes are required.
- **Direct execution**: The delete uses a single `pool.Exec` call without a transaction wrapper, since it is a single-statement operation with no multi-step consistency requirements.
- **Layering follows existing pattern**: SQL query constant in the query layer, a `Delete` function in the pickup model, a `DeleteHandler` in the pickups controller, and the route registered in `setupRoutes`.
- **Authorization**: The existing `PickupsGate.Delete → Owns → HasPickup` chain already correctly authorizes DELETE requests. The shop that owns the pickup is allowed; all other roles (including other shops, consignees, admins) are denied.
- **Error handling**: Standard `util.HandleGuessFromDBOperation` for database errors. Non-existent pickup IDs are caught by the gate (returns 403) before reaching the controller, consistent with the existing PUT behavior.
- **Response format**: `{"success": true, "message": "ponto de entrega deletado com sucesso"}` with HTTP 200, matching the product delete endpoint pattern.
- **No schema migration needed**: The `ON DELETE CASCADE` on the junction table is already present from migration 000041.
- **No gate changes needed**: The gate's `Dispatch` already routes DELETE to `p.Delete`, which calls `p.Owns`, which checks `shopmodel.HasPickup` using the param `:pickup`.

## Testing Decisions

- **Integration tests only**: Since this is a thin CRUD endpoint with no novel business logic, integration tests against the HTTP API are the right level. They test external behavior (status codes, response bodies, cross-shop isolation) without coupling to implementation details.
- **Test file**: New subtests added to the existing `TestPickup` function alongside create, list, and update tests.
- **Test scenarios**:
  - Happy path: create a pickup, delete it, verify it no longer appears in the list.
  - Unauthorized: attempt DELETE without authentication — expect 401.
  - Cross-shop isolation: shop A creates a pickup; shop B attempts to delete it — expect 403.
  - Not found: attempt to delete a non-existent pickup ID — expect 403 (caught by gate, consistent with PUT behavior).

## Out of Scope

- Soft delete or status-based deactivation.
- Cleanup of orphaned address rows.
- Bulk deletion of multiple pickup locations.
- Admin override to delete any shop's pickup location.
- Any change to how Consignments reference Pickup Locations.

## Further Notes

- The `// TODO: View/Delete` comment in the pickup model file is resolved by this feature.
- No changes to the policy bouncer, authorization gate, or database schema are required.

Triage: ready-for-agent





## Summary of Changes

Implemented DELETE /pontos/:pickup endpoint for deleting pickup locations.

### Changes

1. **query/pickup.go**: Added `DeletePickupLocationQuery` SQL constant (`DELETE FROM pickup_locations WHERE id = $1`)
2. **models/pickupmodel/model.go**: Added `Delete(ctx, base, id)` function — single `pool.Exec` call, no transaction wrapper (matching existing pattern)
3. **controllers/pickups/pickups.go**: Added `DeleteHandler` — parses the `:pickup` param, calls `pickupmodel.Delete`, returns `util.OK` with success message
4. **routes.go**: Registered `DELETE :pickup` route under the `/pontos` group
5. **api_pickup_test.go**: Added 5 subtests:
   - `delete_pickup` — happy path: delete and verify removal from list
   - `delete_pickup_unauthorized` — expect 401 without auth
   - `delete_pickup_cross_shop` — shop B cannot delete shop A's pickup (403)
   - `delete_pickup_not_found` — non-existent ID returns 403 (caught by gate)
   - `delete_pickup_bad_param` — non-integer param returns 403 (caught by gate)

No schema migration, gate changes, or policy changes required. The existing `ON DELETE CASCADE` on `consignment_pickup_locations` handles junction-table cleanup automatically.
