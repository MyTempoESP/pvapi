---
# pvapi-lt8e
title: 'Admin: list all saques across shops'
status: completed
type: feature
priority: normal
created_at: 2026-05-29T16:33:15Z
updated_at: 2026-05-29T16:38:28Z
---

## Problem Statement

Administrators have no way to view saques (withdrawals) across all shops. The existing saque endpoints live under `/estabelecimentos/saques` and are scoped to the authenticated shop's ID — an admin logged in with a shop role can only see that shop's saques. There is no endpoint that lets an admin see every saque in the system.

## Solution

Add a `GET /admin/saques` endpoint, protected by the existing admin authorization guard, that returns a list of all saques across all shops ordered by most recent first. Reuses the existing `SaqueResource` shape — same fields, same structure — so no new response format.

## User Stories

1. As an admin, I want to list all saques across all shops, so that I can monitor withdrawal activity system-wide.
2. As an admin, I want to see the same saque fields (description, asked_value, user_name, shop_name, bank_name, pix, status, created_at) that shop users see, so that I have consistent data.
3. As an admin, I want saques ordered by most recent first, so that I see the latest activity first.
4. As a non-admin user (shop, consignee, client), I want to be blocked from the admin saques endpoint, so that only admins can see cross-shop data.

## Implementation Decisions

* **Endpoint**: `GET /admin/saques` — list-only, no filtering or pagination for the initial implementation.
* **Authorization**: inherited from the existing `/admin` guard in the policy bouncer, which requires `user.Role == "admin"`.
* **Query**: new `ListAllSaquesQuery` — identical SELECT + JOINs as the existing `ListSaquesByShopQuery` but without the `WHERE sq.shop_id = $1` clause.
* **Model**: new `ListAll(ctx, base) ([]SaqueRow, error)` in `saquemodel`, reusing `SaqueRow` and `ToResource()`.
* **Controller**: new `AdminListHandler(pool)` in `controllers/saque`, following the same pattern as the existing `ListHandler`.
* **Routing**: added inside the existing `admin` route group alongside the existing moderation routes.
* **Response**: `{"saques": [...]}` — same wrapper key as the shop-scoped list, same `SaqueResource` fields.

## Testing Decisions

* Integration tests in `api_saque_test.go` following the existing harness pattern.
* Test cases:
  * Admin can list saques from multiple shops.
  * Response contains expected fields (description, status, shop_name, bank_name, pix, etc.).
  * Unauthorized user (no auth) gets 401.
  * Non-admin authenticated user (shop role) gets blocked by the admin guard.
* Prior art: existing `TestSaque` integration test and `api_harness_test.go` helpers.

## Out of Scope

* Filtering by status, shop, or date range.
* Pagination.
* Individual saque view/update/delete under `/admin/saques` (shop-scoped endpoints already handle that).
* New resource types or response shapes.

## Further Notes

The admin guard in `controllers/policy/policy.go` already registers `/admin` with `gate.NewAdmin`, so no policy changes are needed. The new route is automatically covered.

Triage: ready-for-agent


## Summary of Changes

Implemented `GET /admin/saques` endpoint that lists all saques across all shops.

### Changes

- **query/saque.go**: Added `ListAllSaquesQuery` — same SELECT + JOINs as `ListSaquesByShopQuery`, but without the `WHERE sq.shop_id = $1` clause. Saques are ordered by most recent first.
- **models/saquemodel/model.go**: Added `ListAll(ctx, base) ([]SaqueRow, error)` — reuses `SaqueRow` and `ToResource()` from existing model.
- **controllers/saque/saque.go**: Added `AdminListHandler(pool)` — follows same pattern as existing `ListHandler`, using `util.ResourceCollection` and `util.WrapLabeled`.
- **routes.go**: Added `admin.GET("/saques", saque.AdminListHandler(pool))` inside the existing `/admin` route group. No policy changes needed — the existing admin guard already covers it.
- **api_saque_test.go**: Added `TestAdminListSaques` with 3 sub-tests:
  - `admin_list_all`: Admin can list saques from multiple shops, verifying fields (description, status, bank_name, shop_name, user_name, asked_value, created_at)
  - `admin_list_unauthorized`: Unauthenticated user gets 401
  - `admin_list_blocked_for_shop\]: Non-admin authenticated user (shop role) gets 403
