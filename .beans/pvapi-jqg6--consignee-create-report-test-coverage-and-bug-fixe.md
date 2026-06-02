---
# pvapi-jqg6
title: Consignee Create Report — test coverage and bug fixes
status: todo
type: feature
priority: normal
created_at: 2026-06-02T22:17:46Z
updated_at: 2026-06-02T22:17:50Z
---

## Problem Statement

Consignees (operators at pickup locations) need to create occurrence reports on products and consignments they handle — but the `POST /ocorrencias` endpoint had no test coverage for consignee-authenticated requests. Two controller bugs also slipped in during implementation:

1. The consignee branch of `CreateHandler` did not return after writing the response, causing a fall-through that attempted a second DB insert and a double response write.
2. Scope validation (`validateScope`) only ran for shops, not consignees — a consignee could create reports on products/consignments they had no relationship to.

## Solution

Fix the two controller bugs and add comprehensive test coverage for consignee report creation inside `TestReport`, covering both success and rejection paths across product and consignment scopes.

## User Stories

1. As a consignee, I want to create an occurrence report on a product I have access to, so that I can flag issues like missing parts or damage.
2. As a consignee, I want to create an occurrence report on a consignment I'm assigned to, so that I can report problems with the shipment I'm managing.
3. As a consignee, I want the system to reject my report if I reference a product that doesn't exist, so that data integrity is maintained.
4. As a consignee, I want the system to reject my report if I reference a product belonging to a different shop, so that I can't interfere with other shops' inventory.
5. As a consignee, I want the system to reject my report if I reference a consignment that doesn't exist, so that data integrity is maintained.
6. As a consignee, I want the system to reject my report if I reference a consignment assigned to a different consignee, so that I can't report on shipments I'm not responsible for.
7. As a consignee, I want my created reports to appear in my own report listing, so that I can track issues I've flagged.
8. As a shop owner, I want consignee-created reports on my products and consignments to appear in my report listing, so that I can see all issues affecting my inventory.

## Implementation Decisions

1. **Bug fix — early return**: `CreateHandler` in the report controller now returns immediately after `consigneeReportCreationHandler` completes, preventing the shop-only `report.Prepare()` / `report.Save()` path from executing for consignees.

2. **Bug fix — scope validation ordering**: `validateScope` is called before the role switch, so scope ownership checks run for both shops and consignees. The existing `consigneemodel.HasProduct` and `consigneemodel.HasConsignment` queries handle consignee-side validation.

3. **Test location**: All new subtests live inside the existing `TestReport` top-level test function, which already has consignee session setup and relevant fixtures.

4. **Test fixtures for cross-ownership rejection**: A second shop with its own product, and a second consignment assigned to a different monitor, are created in the test setup to validate ownership boundaries.

## Testing Decisions

- **Good tests test external behavior**: Each subtest makes an HTTP request to `/ocorrencias` and asserts the status code; listing subtests verify that reports appear in both shop and consignee list views. No internal model/DB assertions.
- **Prior art**: The existing shop-side subtests in `TestReport` (`create_product_scope_report`, `reject_nonexistent_product`, `list_shop_reports_and_monitor_list`, etc.) establish the pattern for consignee-side coverage.
- **Test structure**: Subtests follow the same naming convention as existing ones (`consignee_create_product_report`, `consignee_reject_product_nonexistent`, etc.) and use the same `reportFixture` / `doJSON` / `assertStatus` helpers.

## Out of Scope

- Refactoring the report controller into separate shop and consignee handlers.
- Adding update or delete test coverage for consignee-created reports.
- Changing the error message that says "does not belong to this shop" when the caller is a consignee.
- Adding tests for edge cases like empty request bodies or malformed JSON payloads.

## Further Notes

- The report controller still has legacy handling where `occurrenceScope: "shipment"` is mapped to `"consignment"` — this is noted in a TODO comment and not addressed here.
- The `ReportWithConsigneeIDBlueprint` persists `consignee_id` directly on the report row, while the shop-side `ReportBlueprint` uses the monitor name string from the request body — this asymmetry is pre-existing and not changed.

Triage: ready-for-agent
