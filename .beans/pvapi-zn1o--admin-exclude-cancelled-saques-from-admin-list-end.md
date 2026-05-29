---
# pvapi-zn1o
title: 'Admin: exclude cancelled saques from admin list endpoint'
status: todo
type: feature
priority: normal
created_at: 2026-05-29T20:11:57Z
updated_at: 2026-05-29T20:12:01Z
---

## Problem Statement

The `GET /admin/saques` endpoint returns all saques across all shops, including cancelled ones. Cancelled saques represent withdrawals the shop itself decided not to pursue — they are noise in the admin moderation view. An admin browsing the saque list for moderation purposes (approve/reject) has to visually skip over cancelled entries, which adds friction as the system grows.

## Solution

Filter out saques with `status = 'cancelled'` from the admin list endpoint. The shop-scoped `GET /estabelecimentos/saques` endpoint remains unchanged so shops can still see their own cancelled saques.

## User Stories

1. As an admin, I want to see only non-cancelled saques when listing all saques across shops, so that I can focus on pending/rejected/approved saques that require my attention.
2. As an admin, I want cancelled saques excluded from the list automatically so that I don't have to mentally filter them out while scanning for pending items to moderate.
3. As an admin, I want newly added saque statuses (beyond the current set) to automatically appear in the admin list if they are not the terminal "cancelled" state, so that the filter doesn't silently exclude future statuses.
4. As a shop user, I want my cancelled saques to still appear in my shop-scoped saque list, so that I retain full visibility into my own withdrawal history.
5. As a developer, I want the filter to reuse the existing `SaqueRow` struct and `SaqueResource` serialization so that the admin list response shape is unchanged.

## Implementation Decisions

* **Endpoint**: `GET /admin/saques` — filter applied at the SQL level.
* **Shop endpoint unchanged**: `GET /estabelecimentos/saques` continues to return all saques including cancelled.
* **New query**: `ListAdminSaquesQuery` — identical SELECT and JOINs as `ListAllSaquesQuery` but adds `WHERE sq.status != 'cancelled'`.
* **New model function**: `ListForAdmin(ctx, base) ([]SaqueRow, error)` in `saquemodel`, reusing existing `SaqueRow` struct.
* **Controller update**: `AdminListHandler` calls `saquemodel.ListForAdmin()` instead of `saquemodel.ListAll()`.
* **`ListAllSaquesQuery` and `ListAll()` preserved**: kept unchanged for potential future use (it has no other callers today).
* **Filter uses exclusion (`!= 'cancelled'`) rather than inclusion (`IN ('pending', 'approved', 'rejected')`)**: if a new status is added, it will automatically appear in the admin list without a query change.
* **No new request parameters**: no query string filters, no pagination — same shape as before.

## Testing Decisions

* Integration tests in `api_saque_test.go` following the existing harness pattern.
* New sub-test in `TestAdminListSaques`:
  * Create a saque, cancel it via the shop endpoint, then verify it is absent from the admin list.
  * Create another pending saque to confirm the list still returns non-cancelled entries.
* Prior art: existing `TestAdminListSaques` test and `api_harness_test.go` helpers.

## Out of Scope

* Filtering rejected or approved saques from the admin list — only cancelled is excluded.
* Filtering saques by shop, date range, or other criteria.
* Pagination on the admin saques list.
* Changing the shop-scoped list endpoint (`GET /estabelecimentos/saques`) behavior.
* Adding a separate endpoint for "cancelled saques" or any archive/trash feature.

## Further Notes

The existing `UpdateSaqueRequest` has binding `oneof=cancelled` — the shop endpoint only allows shops to set status to `cancelled`. Admin approval and rejection go through separate moderation endpoints (`/admin/moderacao/saques/:id/aprovar` and `/.../rejeitar`), which check `status = 'pending'` before acting. This filter therefore removes from the admin list exactly the saques that an admin can never act on again (cancelled is terminal, while rejected leaves the shop's balance freed and the record for audit).

Triage: ready-for-agent
