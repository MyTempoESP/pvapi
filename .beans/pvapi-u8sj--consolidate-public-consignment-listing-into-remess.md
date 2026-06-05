---
# pvapi-u8sj
title: Consolidate public consignment listing into /remessas-publicas with shop filter
status: done
type: feature
priority: normal
created_at: 2026-06-05T01:49:42Z
updated_at: 2026-06-05T01:49:52Z
---

## Problem Statement

The public integration API currently exposes two endpoints for listing consignments: `GET /remessas-publicas` (all active consignments) and `GET /estabelecimentos/{shop}/remessas` (consignments by shop slug). Having two endpoints is unnecessary surface area. The `/estabelecimentos/{shop}/remessas` route uses a shop slug which is inconsistent with the rest of the integration API (which uses integer IDs). The `ConsignmentResource` also omits the shop that owns each consignment, forcing integrators to cross-reference against the establishments listing.

## Solution

Drop `/estabelecimentos/{shop}/remessas` from public documentation (keep the route functional for backward compatibility). Extend `/remessas-publicas` with a `?loja` query parameter (multi-value, integer IDs) matching the pattern already used by `GET /produtos`. Include a `shop` object (`{ id, name }`) in each consignment resource so integrators know which shop owns the consignment without extra lookups.

## User Stories

1. As an integrator, I want to list active consignments via a single `/remessas-publicas` endpoint, so that I do not need to learn two different endpoints for the same data.
2. As an integrator, I want to filter consignments by shop ID(s) using `?loja=1&loja=2`, so that I can get consignments for specific shops in one request.
3. As an integrator, I want each consignment in the listing to include a `shop` object with `id` and `name`, so that I know which shop owns the consignment without querying `/estabelecimentos`.
4. As an integrator filtering by shop, I want the response shape (including `shop`) to be identical to the unfiltered listing, so that client code works uniformly.
5. As an integrator, I want the `?loja` filter to support multiple values just like `GET /produtos`, so that the API feels consistent across endpoints.
6. As a platform operator, I want the old `/estabelecimentos/{shop}/remessas` route to keep working for existing integrations, so that nothing breaks during the transition.

## Implementation Decisions

- **New resource type**: `PublicConsignmentResource` contains all fields from `ConsignmentResource` plus a `shop` object (`{ id: int, name: string }`). Only `GET /remessas-publicas` uses this resource; other consignment endpoints keep `ConsignmentResource` unchanged.
- **Filter parameter**: `?loja` accepts multiple integer values via `c.Request.URL.Query()["loja"]`, parsed with the same `parseQueryInts` helper used by the product listing handler. When absent, returns all active consignments (current behavior).
- **Query strategy**: When `?loja` is present, build a squirrel query with `WHERE shop_id IN (...)` against the `consignments_listing_view`, filtered to active consignments only. Follow the pattern from `productmodel.PublicIndex` (squirrel with `sq.Eq` for multi-value IN clause). When `?loja` is absent, use the existing `ListAllActiveConsignmentsQuery`.
- **Model layer**: Add `PublicListFiltered(ctx, base, shopIDs []int) ([]Consignment, error)` to `consignmentmodel`. The existing `PublicList` remains for the no-filter case.
- **Handler**: Modify `PublicListHandler` in `controllers/consignment/consignment.go` to parse `?loja`, delegate to the appropriate model function, and map results through the new `PublicConsignmentResource`.
- **Docs only — no route removal**: The route `GET /estabelecimentos/:shop/remessas` stays in `routes.go` and continues to work. Only `Dev/llms.txt` and `Dev/requests_implemented/api-integrations.yaml` are updated. `Dev/spec.html` will be re-rendered by the team separately.
- **Schema name**: Use `RecursoRemessaPublica` in the OpenAPI spec to distinguish from `RecursoRemessa`.

## Testing Decisions

- Follow the existing test pattern in `api_consignment_test.go` (end-to-end HTTP tests via `newAPIHarness`).
- Test 1: `GET /remessas-publicas` without filter returns all active consignments, each with `shop: { id, name }`.
- Test 2: `GET /remessas-publicas?loja=X` returns only consignments belonging to shop X.
- Test 3: `GET /remessas-publicas?loja=X&loja=Y` returns consignments for both shops.
- Test 4: `GET /remessas-publicas?loja=99999` returns empty array (nonexistent shop).
- Test 5: Existing `GET /estabelecimentos/{slug}/remessas` tests continue to pass (route unchanged).
- Existing `GET /remessas-publicas` tests will need the response assertion updated to include the new `shop` field.

## Out of Scope

- Changing `/estabelecimentos/{shop}/produtos` (uses slug pattern but is not part of this change).
- Removing or modifying the route in `routes.go`.
- Updating `Dev/spec.html` (team will re-render from the updated YAML).
- Adding shop info to non-public consignment endpoints (`/remessas`, `/monitores/me/remessas`).
- Adding `?remessa` filter to `/remessas-publicas` (only `?loja`).

## Further Notes

- The `Consignment` struct already carries `shop_id` and `shop_name` from the view query. No schema changes needed — just expose it in the resource.
- The `?loja` parameter name uses "loja" (the `/produtos` convention), not "shop", for consistency with the existing public API surface.

Triage: ready-for-agent
