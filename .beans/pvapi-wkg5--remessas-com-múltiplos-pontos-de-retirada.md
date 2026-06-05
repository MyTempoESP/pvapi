---
# pvapi-wkg5
title: Remessas com múltiplos pontos de retirada
status: todo
type: feature
priority: normal
created_at: 2026-06-05T02:54:11Z
updated_at: 2026-06-05T02:54:16Z
---

## Problem Statement

Atualmente, uma remessa (consignment) está vinculada a exatamente um ponto de retirada (pickup location) via FK `pickup_location_id`. Isso obriga shops a criarem múltiplas remessas idênticas quando querem disponibilizar os mesmos produtos em pontos de retirada diferentes, gerando duplicação operacional e overhead de gerenciamento.

## Solution

Permitir que uma remessa aceite múltiplos pontos de retirada. O campo `destinationId` na criação da remessa passa de `int` para `[]int`. Endpoints separados permitem adicionar/remover pontos de uma remessa existente. No checkout, o cliente escolhe em qual ponto específico retirar cada pacote. A listagem de produtos (catalog view) reflete cada pacote × ponto como entradas independentes.

## User Stories

1. As a shop owner, I want to create a consignment with multiple pickup locations, so that I don't need to duplicate consignments for each pickup point.
2. As a shop owner, I want to add a pickup location to an existing consignment, so that I can expand availability without recreating the consignment.
3. As a shop owner, I want to remove a pickup location from a consignment, so that I can adjust availability when a pickup point closes.
4. As a shop owner, I want the same set of products/packages in the consignment to be automatically available at all attached pickup locations, so that I don't manage per-location stock separately.
5. As a client browsing the catalog, I want to see all pickup location options a product is available at, so that I can choose the most convenient one.
6. As a client at checkout, I want to select which pickup location to use for each item in my purchase, so that items can be collected at different points if needed.
7. As a client, I want to receive a confirmation email grouped by pickup location, showing which items to collect at each point.
8. As a client, I want to see my order overview with the chosen pickup location for each purchased item, so that I know where to go.
9. As a consignee (monitor), I want to see consignments with their full list of pickup locations, so that I know all available destinations.
10. As an unauthenticated visitor, I want to view public consignments with their pickup locations, so that I know where products are available.
11. As a shop owner, I want the system to validate that pickup locations I add to a consignment actually belong to my shop, so that I don't misconfigure by accident.
12. As a shop owner, I want the system to prevent adding the same pickup location twice to the same consignment, so that I don't accidentally duplicate entries.

## Implementation Decisions

### Schema change

A new join table `consignment_pickup_locations(consignment_id, pickup_location_id)` replaces the single `consignments.pickup_location_id` FK. Both columns cascade on delete. The pair has a UNIQUE constraint to prevent duplicates. Existing rows are migrated into the join table before the old column is dropped.

The `consignments_listing_view` drops its `LEFT JOIN pickup_locations` clause — pickup resolution moves to Go code via batch fetch.

### API contract

**Create consignment** — `destinationId` becomes `[]int`. The handler pre-validates that all IDs belong to the shop before insertion. Example body shape:

```json
{"destinationId": [1, 2, 3], "monitorId": 1, "status": "RASCUNHO", "monitorProfit": 0}
```

**Update consignment** — the `PUT /remessas/:consignment` endpoint drops `destinationId` entirely. Pickup management moves to dedicated endpoints.

**New endpoints:**
- `POST /remessas/:consignment/pontos` — add a pickup. Body: `{"destinationId": 5}`
- `DELETE /remessas/:consignment/pontos/:pickup` — remove a pickup

**Consignment resource shape** — `destination` (string) and `PickupName` are replaced by a `destinations` array:

```json
{"destinations": [{"id": 1, "name": "Shopping Center"}, {"id": 2, "name": "Praça Central"}]}
```

### Batch pickup resolution

A new helper `BatchFetchConsignmentPickups(ctx, base, consignmentIDs)` returns `map[consignmentID][]Pickup`. Every listing/view handler that returns consignment resources must call this after scanning, setting each consignment's `Pickups` field before `ToResource()`. This follows the existing `BatchFetchPickups` pattern used for products.

### Product pickups (catalog)

The `PackagesByProductIDs` query joins `packages → consignments_active → consignment_pickup_locations → pickup_locations` instead of the old direct FK path. A package in a consignment with 3 pickup locations now produces 3 `ProductPickupResource` entries — one per pickup. The consumer (`BatchFetchPickups`) groups these by product ID as before.

### Checkout flow

`ConfirmItem` gains `pickup_location_id`. At checkout, the client chooses the specific pickup location per package. The `package_reserved` table gains a `pickup_location_id` column to record this choice.

The email resolution query (`ResolvePickupEmailDataQuery`) and client overview query (`ClientOverviewPackagesQuery`) resolve pickup info directly from `package_reserved.pickup_location_id` instead of joining through consignments.

### Model struct changes

- `ConsignmentBlueprint`: `PickupID int` → `PickupIDs []int`
- `ConsignmentUpdateBlueprint`: drops `PickupID` field
- `CreateConsignmentRequest`: `PickupID int` → `PickupIDs []int` (JSON: `destinationId`)
- `UpdateConsignmentRequest`: drops `PickupID` field
- `Consignment` / `ConsignmentWithStats`: remove embedded `Pickup`, add `Pickups []Pickup` with `db:\\"-\\"`
- `ConsignmentResource` / `PublicConsignmentResource` / `ConsignmentWithStatsResource`: replace single destination field with `Destinations []PickupResource`

### New action types

- `AddPickupAction{ConsignmentID, PickupID}` with `Exec` that inserts into the join table
- `RemovePickupAction{ConsignmentID, PickupID}` with `Exec` that deletes from the join table

## Testing Decisions

All existing integration tests that exercise consignment or pickup flows must be fixed to match the new schema and API shapes. Affected test files include the consignment, checkout (client and shop), product, monitor, and client test suites.

Tests follow the existing patterns in the codebase: integration tests against a real temporary PostgreSQL database, exercising HTTP endpoints and verifying response shapes, status codes, and business logic outcomes.

## Out of Scope

- Backward compatibility with the old single-`destinationId` request format
- Product-specific pickup overrides within a consignment (all products share the consignment's pickup set)
- Changing how `package_reserved` financial calculations (tax, commission) work — only the pickup location resolution path changes
- UI/frontend changes

## Further Notes

- The migration is a clean break — no backwards compat shim. Existing consignments get their single pickup migrated to the join table.
- Pickup add/remove operations have no status guard — they work on any consignment regardless of status (RASCUNHO, active, etc.).
- The `consignments_listing_view` becomes lighter by dropping the pickup join, matching the existing pattern of batch-resolving relational data in Go code.

Triage: ready-for-agent
