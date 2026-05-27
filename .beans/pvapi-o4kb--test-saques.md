---
# pvapi-o4kb
title: Test saques
status: completed
type: task
priority: normal
created_at: 2026-05-27T06:06:29Z
updated_at: 2026-05-27T06:06:59Z
---

Run the saque (withdrawal) tests to verify all endpoints work correctly

## Summary of Changes

Ran 20 saque test cases from `TestSaque` in `api_saque_test.go`. All passed:
- CRUD operations: create, list, view, update status, delete — including auth checks and edge cases
- Balance validation: negative value, zero value, exceeds balance, exact balance, pending saques reduce available balance
- Checkout report correctly reflects pending saque deductions

All tests passed in 10.5s.
