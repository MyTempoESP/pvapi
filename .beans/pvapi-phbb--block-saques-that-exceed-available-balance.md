---
# pvapi-phbb
title: Block saques that exceed available balance
status: completed
type: feature
priority: normal
created_at: 2026-05-27T04:09:18Z
updated_at: 2026-05-27T05:17:10Z
---

## Problem Statement

Shop owners can create withdrawal requests (saques) for any amount, even when the requested value exceeds their available earnings. The system also displays an incomplete balance — the checkout report's `actual` field subtracts only approved saques, not pending ones. This means:

- A shop with R$100 in earnings and R$80 in pending saques sees `actual: 100` instead of the true available `20`
- A shop can create a saque that mathematically cannot be fulfilled

## Solution

Block saque creation when `asked_value` exceeds the shop's available balance. Update the checkout report totals so pending saques are included in the `actual` deduction and the `saques` field reflects all non-terminal saques (approved + pending).

## User Stories

1. As a shop owner, I want the system to reject a saque that exceeds my available balance, so I don't create impossible withdrawal requests.
2. As a shop owner, I want to see my true available balance on the checkout report, including pending saques in the deduction, so I can make informed withdrawal decisions.
3. As a shop owner, I want the checkout report's `saques` total to reflect both approved and pending saques, so I see the full picture of my locked-up funds.
4. As a shop owner, I want an error response that tells me both the asked value and my available balance when a saque is rejected, so I know how much I can withdraw without extra API calls.
5. As a shop owner, I want saques with negative or zero values to be rejected, so I can't accidentally create invalid requests.
6. As a shop owner, I want a saque that equals exactly my available balance to be accepted, so I can withdraw all my earnings at once.
7. As a shop owner, I want concurrent saque creation attempts to be serialized per shop, so two people on the same team don't accidentally overdraw.
8. As an admin, I don't want the approval flow to change — saques are blocked at creation, not at approval time.

## Implementation Decisions

### New module: `models/financemodel/`

A new leaf package encapsulating shop balance logic. Two exported functions:

- **`AvailableBalance(ctx, base, shopID) → (decimal.Decimal, error)`**: Locks the shop row (`SELECT ... FOR UPDATE`), runs two SUM queries (checkout earnings and saque totals by status), returns the available balance. Takes a `db.Queryable` so it works inside or outside transactions.

- **`ComputeSaqueTotals(rows []saquemodel.SaqueRow) → SaqueTotals`**: Pure function that aggregates saque rows into approved and pending sums. Used by `BuildCheckoutReport` which already has all saque rows in memory.

### New file: `query/finance.go`

Two raw SQL queries owned by the finance module:
- `SUM(asked_value)` from saques grouped by status for a given shop
- `SUM(value - tax - commission)` from checkouts via package_reserved for a given shop

### Modified: `controllers/saque/saque.go` — `CreateHandler`

The creation flow becomes:

1. Validate `asked_value > 0` (negative/zero → 400 with message `"valor do saque deve ser positivo"`)
2. Validate `HasBankAccount` (existing, stays)
3. Begin transaction
4. Call `financemodel.AvailableBalance(ctx, tx, shopID)` — acquires row lock + computes balance
5. If `asked_value > available_balance` → 400 with labeled response:
   ```json
   {"message": "saldo insuficiente para saque", "saque": {"asked_value": "...", "available_balance": "..."}}
   ```
6. Call `saquemodel.Create(ctx, tx, ...)` — INSERT within same transaction
7. Commit

### Modified: `models/shopmodel/checkout.go` — `BuildCheckoutReport`

- Replace the inline saque aggregation loop with `financemodel.ComputeSaqueTotals(saqueRows)`
- `Totals.Saques` = approved + pending combined
- `Totals.Actual` = total − taxes − commissions − `Totals.Saques`
- Remove the separate `saquesTotal` variable

### Modified: `models/shopmodel/checkout_resources.go`

- No struct changes needed. The `TotalsResource.Saques` field now carries the combined approved+pending sum. The `json:"saques"` tag stays.

### Race condition prevention

A `SELECT ... FOR UPDATE` on the shop row serializes concurrent saque creation for the same shop. The lock is held for the duration of the transaction (balance read + INSERT), released on commit/rollback. Concurrent requests block until the first commits, then see the updated balance.

### Negative balance handling

No special code. If balance is negative (shouldn't happen in normal use), `asked_value > balance` naturally blocks all positive saques.

## Testing Decisions

### Test style

Follow the existing pattern in `api_saque_test.go`: black-box HTTP tests using the `apiHarness` fixture helpers. No unit tests for internal functions — test through the API contract.

### New test cases to add to `TestSaque`

1. **create_exceeds_balance** — shop has checkout earnings (~3.00 after tax/commission), attempts saque of 150.00 → 400 with balance info in response
2. **create_within_balance** — saque value ≤ available balance → 200 OK
3. **create_exact_balance** — saque equals available balance exactly → 200 OK
4. **create_negative_value** — asked_value is negative → 400
5. **create_zero_value** — asked_value is 0 → 400
6. **create_pending_reduces_balance** — create saque A (small), create saque B (small), verify third saque is blocked because balance exhausted
7. **checkout_report_reflects_pending** — GET /estabelecimentos/checkouts: verify `totals.saques` and `totals.actual` include pending saques

### Test setup

Tests that need earnings will use the full checkout flow (create shop → create product → approve → create consignment → add product → create cart item → confirm checkout) to create actual revenue, then test saque creation against it. This mirrors how `TestCartAndCheckout` sets up data.

## Out of Scope

- **Approval-time balance re-validation**: Saques blocked at creation only. Admin approval is the human gate.
- **Gate refactoring**: `HasBankAccount` stays in the controller for now.
- **Concurrent checkout + saque race conditions**: The lock only covers concurrent saques. A checkout finishing between balance-read and saque-insert is not locked. This is acceptable — the window is small and unlikely.
- **Value editing on existing saques**: Already impossible (UpdateSaqueRequest only has Status).

## Further Notes

- The `SELECT ... FOR UPDATE` locks the shops row. Any future balance-mutating paths (e.g., refunds, reversals) should also lock the shop row in the same pattern.
- The `AvailableBalance` function is designed to be reusable for any endpoint that needs to check shop balance before performing a financial operation.

Triage: ready-for-agent

## Summary of Changes

### New files

- **query/finance.go**: SQL queries for balance computation — `SaqueTotalsByShopQuery`, `CheckoutEarningsByShopQuery`, `LockShopRowQuery`
- **models/financemodel/balance.go**: New package with `AvailableBalance` (locks shop row via `SELECT ... FOR UPDATE`, sums earnings and non-terminal saques) and `ComputeSaqueTotals` (pure function that aggregates approved + pending saque sums)

### Modified files

- **controllers/saque/saque.go**: `CreateHandler` now validates `asked_value > 0`, wraps creation in a transaction, calls `financemodel.AvailableBalance`, and rejects with 400 + labeled balance info when balance is insufficient
- **models/shopmodel/checkout.go**: `BuildCheckoutReport` uses `financemodel.ComputeSaqueTotals` instead of inline loop; `Saques` field now includes pending + approved (combined), and `Actual` deducts both
- **api_saque_test.go**: 7 new test cases (`create_exact_balance`, `create_negative_value`, `create_zero_value`, `create_exceeds_balance`, `create_pending_reduces_balance`, `checkout_report_reflects_pending`); existing tests updated to use small saque amounts compatible with balance validation
- **api_fixtures_test.go**: Added `generateEarnings` helper that runs the full checkout flow with PERCENT commission (1%) to produce positive net earnings for balance tests
