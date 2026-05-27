---
# pvapi-7xnw
title: GET /monitores/me/remessas — consignee-gated remessa listing
status: todo
type: feature
priority: normal
created_at: 2026-05-27T23:09:53Z
updated_at: 2026-05-27T23:09:58Z
---

## Problem Statement

Consignees (monitores) have no way to list the remessas (consignments) they participate in. Currently, only shops can list their own remessas via `GET /remessas` (filtered by shop ID). A consignee logging in has no endpoint to see which remessas they are assigned to.

## Solution

A new authenticated endpoint `GET /monitores/me/remessas` that returns all active remessas where the authenticated consignee is the assigned monitor. The endpoint is gated to consignee-role users only (shops and admins are rejected).

## User Stories

1. As a consignee (monitor), I want to list all remessas I participate in, so that I can see my workload and plan my deliveries.
2. As a consignee, I want the list to exclude archived remessas, so that I only see active work.
3. As a consignee with no remessas, I want to receive an empty list (not an error), so that the UI can handle the zero-state gracefully.
4. As a shop user, I want to be blocked from accessing the consignee remessa list, so that data boundaries between roles are enforced.
5. As an unauthenticated user, I want to be rejected with 401, so that only authorized consignees can see remessa assignments.
6. As a consignee, I want each remessa in the list to include its ID, status, commission info, pickup location name, and shop name, so that I have enough context to act on each remessa.

## Implementation Decisions

- **Authorization**: Reuses the existing ConsigneeGate. The gate already has a `Me()` method that checks `role == consignee` and the `Dispatch` already routes `/monitores/me` prefix to it. No gate changes needed.
- **Identity**: The consignee identity is derived from `AuthUser.ActiveRoleID`, which for a consignee user holds their consignee entity ID — same pattern used by shops for their shop ID.
- **Query**: A new SQL query filters `consignments_listing_view` by `consignee_id` and excludes archived remessas via the `consignments_active` subquery — mirroring the existing `ListConsignmentsByShopQuery` pattern.
- **Response format**: Reuses the existing `ConsignmentResource` shape (via `ToResource()`), wrapped with `util.Wrap` — identical format to `GET /remessas` and `GET /remessas-publicas`.
- **No pagination or filters**: YAGNI. The existing shop listing also has none. Can be added later if needed.
- **Handler placement**: Lives in the consignment controller package since it returns consignment resources, following the pattern of other consignment handlers.
- **Route registration**: A new `/monitores/me` subgroup under the existing `/monitores` group, with a single `GET /remessas` route.

## Testing Decisions

- Tests live in the existing monitor test file, since the endpoint is under `/monitores/me`.
- Three test cases:
  1. **Happy path**: Create consignee, create remessa with that consignee, login as consignee, hit endpoint → 200 with remessa in list
  2. **Empty list**: Consignee with zero remessas → 200 with empty array
  3. **Authorization rejection**: Shop user hitting endpoint → 403 (gate rejects before handler runs)
- Tests are integration-style, using the existing `apiHarness` and fixture helpers (`createMonitor`, `createConsignment`, `login`, `doRequest`). Only external HTTP behavior is tested — no mocking of internal layers.

## Out of Scope

- Pagination, filtering, or sorting of remessa listings
- Any UI/design work
- Modifications to the ConsigneeGate authorization logic
- Public (unauthenticated) access to consignee remessas
- Cross-consignee visibility (a consignee can only see their own remessas)

## Further Notes

The existing public remessa listing (`GET /remessas-publicas`) and shop-scoped listing (`GET /estabelecimentos/:shop/remessas`) serve different audiences. This endpoint completes the triad by giving consignees their own scoped view.

Triage: ready-for-agent
