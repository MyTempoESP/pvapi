---
# pvapi-89h8
title: 'Consignee consignment listing: add totalProducts and soldProducts fields'
status: completed
type: feature
priority: normal
created_at: 2026-05-29T01:03:19Z
updated_at: 2026-05-29T01:08:51Z
---

## Problem Statement

The consignment listing endpoint for consignees (`GET /monitores/me/remessas`) returns a `ConsignmentResource` object that lacks aggregate product counts. The frontend needs `totalProducts` and `soldProducts` fields to render the shipments list, but they are not present in the current API response.

The frontend expects each shipment item to have this shape:

| Frontend access | Expected field |
|---|---|
| `apiShipment.code` | `code` (string) |
| `apiShipment.id` | `id` (int, parsed from string) |
| `apiShipment.status` | `status` (string) |
| `apiShipment.destination` | `destination` (string) |
| `apiShipment.createdAt` | `createdAt` (ISO date) |
| `apiShipment.totalProducts` | `totalProducts` (int) |
| `apiShipment.soldProducts` | `soldProducts` (int) |

## Solution

Add `totalProducts` and `soldProducts` aggregate fields to the consignee-specific consignment listing endpoint. The endpoint response will include these computed fields so the frontend can render shipment summaries without additional API calls.

Only the consignee listing endpoint changes — the shop listing and public listing endpoints remain unchanged.

## User Stories

1. As a consignee, I want to see how many products are in each of my consignments, so that I can understand the shipment volume at a glance.
2. As a consignee, I want to see how many products have been sold from each of my consignments, so that I can track sales progress without manual counting.
3. As a frontend developer, I want the API response to use camelCase field names (`totalProducts`, `soldProducts`, `createdAt`) matching the TypeScript `Shipment` interface, so that no client-side field renaming or mapping is needed.
4. As a frontend developer, I want the response shape to be a contract I can rely on, containing only the fields the UI consumes, so that extra fields don't clutter the payload.

## Implementation Decisions

- **Scope**: consignee-only. `GET /monitores/me/remessas` changes; `GET /remessas` (shop listing) and `GET /remessas-publicas` (public listing) are unchanged.
- **New structs**: `ConsignmentWithStats` (DB struct with pgx row mapping) and `ConsignmentWithStatsResource` (JSON resource). These are separate from the existing `Consignment` / `ConsignmentResource` structs to avoid affecting the shop-side and public listing endpoints.
- **JSON field naming**: Go struct fields use PascalCase (`TotalProducts`, `SoldProducts`). JSON tags use camelCase exactly matching the frontend contract (`totalProducts`, `soldProducts`, `createdAt`).
- **Resource fields**: Only the 7 fields the frontend consumes: `code`, `id`, `status`, `destination`, `createdAt`, `totalProducts`, `soldProducts`. Extra fields like `commission_type`, `commission_value`, `slug`, `responsibleMonitor` are excluded.
- **SQL strategy**: No database migration. The existing `ListConsignmentsByConsigneeQuery` is updated with COALESCE'd scalar subqueries for the aggregates inline.
  - `totalProducts`: `COALESCE((SELECT SUM(quantity) FROM packages WHERE consignment_id = c.id), 0)`
  - `soldProducts`: `COALESCE((SELECT SUM(pr.quantity) FROM package_reserved pr JOIN packages pkg ON pkg.id = pr.package_id WHERE pkg.consignment_id = c.id), 0)`
- **Model function**: The old `ListByConsignee` function is replaced by a new function that returns `[]ConsignmentWithStats`. The handler is updated to call the new function.
- **No new files**. Changes touch only existing files.

## Testing Decisions

- Extend the existing `consignee_list_remessas_happy_path` test in `api_monitor_test.go` to assert the new `totalProducts` and `soldProducts` fields are present and correct.
- The checkout test (`TestCartAndCheckout`) already sets up the full flow (create consignment, add products, checkout) and can be extended with a sub-test that logs in as the consignee and verifies `soldProducts` reflects the completed checkout.
- Good tests verify external behavior (API response shape and correctness of computed aggregates), not implementation details.
- Prior art: existing `consignee_list_remessas_happy_path` test pattern — create fixtures, call endpoint, assert JSON fields.

## Out of Scope

- Shop-side listing endpoint (`GET /remessas`)
- Public listing endpoint (`GET /remessas-publicas`)
- Changing the `ConsignmentResource` shape used by other endpoints
- Database migrations or view changes
- Pagination or filtering of the consignee listing



## Summary of Changes

**Files changed:**
-  — Added  struct with  and  db tags
-  — Added  (7 fields, camelCase JSON tags matching frontend  interface) and its  method
-  — Updated  with COALESCE'd scalar subqueries for aggregate counts, aliased the view as 
-  — Added  returning 
-  —  now calls  and uses the new resource type
-  — Extended  to add a product and assert , , and camelCase field presence

**What changed in the API response (GET /monitores/me/remessas):**
- Response now uses a slim  shape with only: , uid=1000(rodrigo) gid=1000(rodrigo) groups=10(wheel),18(audio),23(input),27(video),28(netdev),1000(rodrigo), , , , , 
- No more , , ,  in the listing
- JSON field names are camelCase: , , , 
-  = SUM(quantity) from packages for the consignment
-  = SUM(quantity) from package_reserved for the consignment

**Not changed:**
- Shop listing () and public listing () endpoints remain unchanged
- No database migration needed
- Old  kept for backward compatibility

## Summary of Changes

**Files changed:**
- models/consignmentmodel/data.go - Added ConsignmentWithStats struct with total_products and sold_products db tags
- models/consignmentmodel/resources.go - Added ConsignmentWithStatsResource (7 fields, camelCase JSON tags matching frontend Shipment interface) and its ToResource() method
- query/consignment.go - Updated ListConsignmentsByConsigneeQuery with COALESCE scalar subqueries for aggregate counts, aliased the view as c
- models/consignmentmodel/model.go - Added ListByConsigneeWithStats returning []ConsignmentWithStats
- controllers/consignment/consignment.go - ListByConsigneeHandler now calls ListByConsigneeWithStats and uses the new resource type
- api_monitor_test.go - Extended consignee_list_remessas_happy_path to add a product and assert totalProducts=2, soldProducts=0, and camelCase field presence

**API response (GET /monitores/me/remessas):**
- New slim shape: code, id, status, destination, createdAt, totalProducts, soldProducts
- No more commission_type, commission_value, slug, responsibleMonitor
- JSON field names are camelCase
- totalProducts = SUM(quantity) from packages
- soldProducts = SUM(quantity) from package_reserved
