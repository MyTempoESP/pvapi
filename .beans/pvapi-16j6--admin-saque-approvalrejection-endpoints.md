---
# pvapi-16j6
title: Admin saque approval/rejection endpoints
status: completed
type: feature
priority: normal
created_at: 2026-05-29T16:58:25Z
updated_at: 2026-05-29T17:06:19Z
---

## Problem Statement

Shops create saques (withdrawals) that start as `pending`. Currently, the only way to approve or reject a saque is via the shop-level `PUT /estabelecimentos/saques/:saque` endpoint, which any shop user can call. There is no admin-gated approval flow, no rejection reason tracking, and no balance re-verification at approval time. Admins can list all saques via `GET /admin/saques` but cannot act on them.

## Solution

Two new admin-only endpoints under the existing moderation pattern:

- `POST /admin/moderacao/saques/:id/aprovar` — Admin approves a pending saque after re-verifying the shop's available balance
- `POST /admin/moderacao/saques/:id/rejeitar` — Admin rejects a pending saque with a mandatory `motivo` (reason)

The shop-level `PUT` is restricted to only allow `cancelled` status, removing the self-approve/reject backdoor. A new `motivo` column on the `saques` table stores the rejection reason and is exposed in API responses.

## User Stories

1. As an admin, I want to approve a pending saque, so that the shop's withdrawal request is confirmed after I verify it's legitimate.
2. As an admin, I want to reject a pending saque with a reason, so that the shop understands why their withdrawal was denied.
3. As an admin, I want the system to re-check the shop's available balance when I approve, so that I don't approve a saque that would overdraw the shop.
4. As an admin, I want to be blocked from approving or rejecting a saque that is not in `pending` status, so that I can't accidentally double-process a saque.
5. As a shop user, I want to cancel my own pending saques, but I should not be able to self-approve or self-reject them.
6. As a shop user, I want to see the rejection reason (`motivo`) when my saque is rejected, so that I know why it was denied and can fix the issue.
7. As a shop user, I want the rejected saque's value to be freed from my balance calculation, so that I can request a new saque with those funds.

## Implementation Decisions

### Route pattern

Follow the existing moderation convention: `/admin/moderacao/saques/:id/aprovar` and `/admin/moderacao/saques/:id/rejeitar` (POST). Authorization is handled by the existing `AdminGate` which requires `role == "admin"`.

### Database migration

New migration `000034` adds `motivo TEXT NOT NULL DEFAULT ''` to the `saques` table. The column is present on all rows but only populated on rejection.

### Query layer

Two new parameterized queries with a `WHERE status = 'pending'` guard that prevents double-processing:

```
AdminApproveSaqueQuery = UPDATE saques SET status = 'approved' WHERE id = $1 AND status = 'pending'
AdminRejectSaqueQuery  = UPDATE saques SET status = 'rejected', motivo = $2 WHERE id = $1 AND status = 'pending'
```

Zero rows affected means the saque was not in `pending` state.

### Model functions

Two new functions returning `(int64, error)` — the `int64` is `rowsAffected` so the controller can distinguish "not pending" from a database error:

- `AdminApprove(ctx, tx, id)` — conditional status update
- `AdminReject(ctx, tx, id, motivo)` — conditional status update with motivo

### Balance re-check on approval

The approve handler opens a transaction, calls `AvailableBalance(shop)` (which already does `SELECT ... FOR UPDATE` on the shop row), and only commits if `AvailableBalance >= 0`. If the balance is negative, the transaction rolls back and returns 400 `"saldo insuficiente para aprovar saque"`.

Rejection does not require a balance check — the pending funds are freed when status changes to `rejected`.

### Shop-level PUT restriction

`UpdateSaqueRequest.Status` binding validator changes from `oneof=approved rejected cancelled` to `oneof=cancelled`. Shops can only cancel their own pending saques.

### Resource enrichment

`SaqueResource` gains an optional `motivo` field. For non-rejected saques, it's an empty string. For rejected saques, it contains the admin's reason.

### Error responses

| Scenario | Status | Message |
|---|---|---|
| Saque not found | 404 | (standard not found) |
| Saque not pending (approve or reject) | 400 | `saque não está pendente` |
| Balance insufficient (approve) | 400 | `saldo insuficiente para aprovar saque` |
| Missing motivo (reject) | 400 | (binding validation) |
| Unauthorized (non-admin) | 403 | (from AdminGate) |

### Status transition summary

```
pending ──admin approve──▶ approved
pending ──admin reject───▶ rejected  (motivo stored)
pending ──shop cancel────▶ cancelled
```

### Module interfaces

- **Saque model**: `AdminApprove(ctx context.Context, base db.Queryable, id int) (int64, error)` and `AdminReject(ctx context.Context, base db.Queryable, id int, motivo string) (int64, error)`
- **Saque controller**: `AdminApproveHandler(pool *pgxpool.Pool) gin.HandlerFunc` and `AdminRejectHandler(pool *pgxpool.Pool) gin.HandlerFunc`
- **Saque requests**: `AdminRejectSaqueRequest{ Motivo string }` with `binding:"required"`
- **Saque resource**: `Motivo string` added to both `SaqueRow` and `SaqueResource`

## Testing Decisions

Tests follow the existing integration test pattern in `api_saque_test.go` — full HTTP-level tests using the `apiHarness` fixture helper. Tests should verify external behavior only: status codes, response bodies, and side effects (balance changes, motivo presence, status transitions).

A good test in this codebase exercises the full request/response cycle through the Gin router with a real test database, not isolated unit tests.

### Test cases to add

- Admin approves a pending saque → 200, status becomes `approved`, balance unchanged
- Admin rejects a pending saque with motivo → 200, status becomes `rejected`, motivo present in response
- Admin approves with insufficient balance → 400, status stays `pending`
- Admin approves a non-pending saque → 400 "saque não está pendente"
- Admin rejects a non-pending saque → 400 "saque não está pendente"
- Admin approve/reject returns 403 for non-admin users
- Shop user cannot set status to `approved` or `rejected` via PUT (400 validation error)
- Shop user can still cancel a pending saque via PUT (200)
- Rejected saque's value is freed from balance (new saque can be created with those funds)
- Admin list includes motivo field on rejected saques
- Motivo field is empty for non-rejected saques

### Prior art

Existing tests in `api_saque_test.go` already cover saque creation, listing, viewing, status updates, and admin listing. The new tests extend this pattern. The admin moderation pattern is also exercised in product moderation tests.

## Out of Scope

- Pix/bank transfer initiation on approval — handled externally
- Email or push notifications on status changes
- Bulk approve/reject operations
- Admin comment/note beyond the single `motivo` field
- Audit trail beyond the existing `created_at` timestamp
- Rejection reason templates or presets

## Further Notes

- The `motivo` column uses `TEXT NOT NULL DEFAULT ''` to match the existing `description` column convention on the same table.
- Balance math is unaffected by status transitions since `AvailableBalance` already subtracts both `pending` and `approved` saques. The re-check is a safety guard, not a calculation change.
- Diagrams for the saque lifecycle and admin moderation handler flows are available in `docs/diagrams/`.

Triage: ready-for-agent



## Summary of Changes

### Database
- Migration `000034`: Added `motivo TEXT NOT NULL DEFAULT ''` column to `saques` table

### Query layer
- Added `AdminApproveSaqueQuery` and `AdminRejectSaqueQuery` with status='pending' guard
- Added `sq.shop_id` and `sq.motivo` to all SELECT queries

### Model (models/saquemodel/)
- Added `ShopID` and `Motivo` fields to `SaqueRow`
- Added `Motivo` to `SaqueResource` and `ToResource()`
- Added `AdminApprove` and `AdminReject` functions returning (int64, error)
- Restricted `UpdateSaqueRequest.Status` to `oneof=cancelled`
- Added `AdminRejectSaqueRequest` with required motivo

### Controller (controllers/saque/)
- Added `ErrSaqueNotPending` sentinel error
- Added `AdminApproveHandler` — transaction with balance re-check
- Added `AdminRejectHandler` — validates motivo, conditional reject

### Routes
- `POST /admin/moderacao/saques/:id/aprovar` and `POST /admin/moderacao/saques/:id/rejeitar`

### Tests
- Updated shop update_status to cancel_saque; added shop_cannot_approve/reject tests
- Added `TestAdminSaqueModeration` with 11 subtests covering all scenarios from the spec
