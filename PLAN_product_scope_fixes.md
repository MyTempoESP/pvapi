# PLAN: Product-Scoped Reports Fixes

## 1. Simplify `ConsigneeHasProductReportQuery` to use `r.consignee_id` directly

Every report (product or consignment scoped) has a `consignee_id`. The current query jumps through `products.shop_id = consignees.shop_id`, which grants access to *any* consignee in the shop regardless of who created the report. Change the query to match `r.consignee_id = c.id` — identical to `ConsigneeHasReportQuery` — so the two queries collapse into one.

Additionally, add a TODO comment above the consignee resolution CTE in `CreateProductReportQuery` (and `CreateReportQuery`) noting that resolving by `u.name = @consignee_name` is fragile and will be replaced with consignee IDs once the frontend passes them.

## 2. Rename `ConsignmentID` to `OrderID` in `CreateReportRequest` and `ReportBlueprint`

The `ConsignmentID int` field in `CreateReportRequest` (embedded in `ReportBlueprint`) carries a product ID when scope is `"product"`. Rename the field to `OrderID` to match its JSON key `orderId`, and propagate through `ReportBlueprint.Save()` where it's passed as either `consignment_id` or `product_id` depending on scope.

## 3. Eliminate `FindScope` N+1 by embedding scope into authorization queries

Currently `InvolvedIn` and `Owns` in the gate do two queries: `FindScope` then the appropriate `HasXxx` query. Instead, fold the scope check into the authorization query itself using a `CASE` or `OR` condition on `r.scope`. One query returns `hasReport` regardless of scope. The gate methods call the unified query directly, no longer needing `FindScope`. `FindScope` stays available for non-auth use (e.g., resource serialization).

## 4. Validate referenced entity matches scope before report creation

In the controller's `CreateHandler`:
- `scope == "consignment"` → verify `consignment_id` exists and belongs to the requesting shop/consignee's shop using existing `HasConsignment` queries.
- `scope == "product"` → verify `product_id` exists and belongs to the requesting shop using existing `ShopHasProductQuery` / `ConsigneeHasProduct` equivalents.

Return a clean error before hitting the DB foreign key constraint.
