---
# pvapi-7d4k
title: 'packages: add discountType field (percentage / fixed)'
status: todo
type: feature
priority: normal
created_at: 2026-06-19T04:58:30Z
updated_at: 2026-06-19T05:00:01Z
---

## Problem Statement

Shops can only apply percentage-based discounts to packages in consignments. A shop wanting to apply a fixed-value discount (e.g., "R$ 10 off") has no way to express it — the system hardcodes percentage math (`discount / 100`) in checkout calculations and constrains `packages.discount` to 0–100.

## Solution

Add a `discountType` field to packages with values `"percentage"` and `"fixed"`. The checkout math branches on this field: percentage discounts continue to work as `price * (1 - discount/100)`, while fixed discounts work as `price - discount` (capped at zero). Existing packages default to `"percentage"`, preserving current behavior. The field is surfaced on all package read paths (consignment product listings, public consignment-by-code view, QR code entries) and accepted on write paths (adding products to consignments, updating package fields).

Additionally, the `ProductListInConsignmentQuery` currently does not return package-level discount at all — it only returns the product-level discount from the catalog view. This is fixed alongside: the query now returns `packages.discount` and `packages.discount_type`.

## User Stories

1. As a shop, I want to set a fixed-value discount on a package in a consignment, so that I can offer "R$ 5 off" rather than being forced into percentage-based discounts.

2. As a shop, I want the system to default new packages to percentage discounts when I don't specify a type, so that I don't break existing workflows.

3. As a shop, I want to see the discount type on packages I've added to consignments, so that I can verify whether a discount is percentage or fixed.

4. As a shop, I want to update the discount and discount type on an existing package, so that I can correct mistakes or adjust pricing.

5. As a client browsing consignment products, I want to see the package-level discount (not the product-level one), so that the displayed discount reflects what I'll actually pay.

6. As a client viewing my checkout QR codes, I want to see the discount type alongside the discount value, so that I understand how the discount was applied.

7. As a shop receiving checkout revenue, I want fixed discounts to be correctly computed in the tax, commission, and shop-value calculations, so that my earnings are accurate.

8. As a platform operator, I want the system to prevent fixed discounts from exceeding the package price (no negative unit prices), so that financial records remain sane.

9. As a consignee viewing a consignment by code, I want to see discount type on each product entry, so that I can explain pricing to clients at pickup.

10. As a developer, I want the discount calculation logic extracted into a single PostgreSQL function, so that checkout math doesn't duplicate the percentage-vs-fixed branch across multiple expressions.

## Implementation Decisions

### Database

- Add `discount_type text NOT NULL DEFAULT 'percentage'` column to the `packages` table.
- Replace the existing `CHECK (discount >= 0 AND discount <= 100)` with a conditional check: `CHECK (discount >= 0 AND (discount_type = 'fixed' OR discount <= 100))`.
- Create a PostgreSQL function `effective_unit_price(price numeric, discount numeric, discount_type text) RETURNS numeric` that computes:
  - `percentage`: `price * (1 - discount / 100)`
  - `fixed`: `GREATEST(price - discount, 0)`

### SQL queries

- `InsertPackageReservedQuery`: the WITH decrement CTE now `RETURNING ... discount_type`. Tax, commission, and shop_value computations use `effective_unit_price(d.price, d.discount, d.discount_type)` instead of the hardcoded `d.price * (1 - d.discount / 100)` math.
- `ProductAddToConsignmentQuery`: accepts a new `@discount_type` named parameter.
- `ProductListInConsignmentQuery`: now returns `p.discount AS package_discount` and `p.discount_type AS package_discount_type` from the packages table (previously missing).
- `ProductsByConsignmentCode`: now returns `pkg.discount_type AS consignment_discount_type`.
- `EnrichedQRCodeEntriesQuery`: now returns `pkg.discount_type`.

### API contracts

**Request DTOs that gain `discountType`:**

- `AddProductUnit` (POST consignment products): optional `discountType` string field, defaults to `"percentage"` if empty.
- `UpdateConsignmentProductRequest` (PUT consignment product): optional `discountType` and `discount` (the latter was previously not updatable).

**Response DTOs that gain `discountType`:**

- `PackageResource` (inside consignment product listings) gains `discountType` and `discount` fields.
- `ProductConsignmentResource` (public consignment-by-code view) gains `discountType`.
- `QRCodeEntry` (checkout confirmation) gains `discountType`.

**Response DTO that gains package-level `discount` (bugfix):**

- `PackageResource` and `ProductResource` (consignment product listing) now include the package-level `discount` value.

### Validation

- When adding a product to a consignment (`AddProductAction.Exec`): if `discountType` is `"percentage"`, discount must be 0–100 inclusive. If `"fixed"`, discount must be ≥ 0 and ≤ the package price.
- The `discountType` field accepts only `"percentage"` and `"fixed"` (validated via the CHECK constraint and application code).

### Architecture

- The `effective_unit_price` function is a deep module: it encapsulates the branching discount logic behind a simple `(price, discount, discount_type) → numeric` interface. The checkout SQL, tax computations, and any future discount consumers call one function instead of duplicating conditional math.
- `discountType` lives as a column on `packages`, not inside the `attributesvalues` JSONB column, because it affects SQL constraint enforcement and checkout math that the database must compute.

### Go types

- `Package` struct (consignment data) gains `Discount` (`decimal.Decimal`) and `DiscountType` (`string`) fields, mapped to `package_discount` and `package_discount_type` columns.
- `ConsignmentProductRow` gains `ConsignmentDiscountType` (`string`), mapped to `consignment_discount_type`.
- `enrichedEntryRow` (checkout internal) gains `DiscountType` (`string`).
- `UpdateConsignmentProductBlueprint` gains `Discount` and `DiscountType` fields, wired into the update field map.
- All `ToResource()` mappers in consignment and checkout models are updated to wire the new fields.

## Testing Decisions

- **Approach**: Integration tests via the test-code skill. Tests target the public API endpoints that read and write packages.
- **What makes a good test**: only test external behavior. Verify that creating a package with `discountType: "fixed"` sets it correctly; that listing consignment products returns the correct discount and discount type; that checkout calculations produce correct shop/tax/commission values for both discount types.
- **Modules targeted**: consignment product CRUD endpoints, checkout confirmation endpoint, public consignment-by-code endpoint.
- **Prior art**: existing integration tests in `controllers/consignment/` and `controllers/checkout/` follow the same pattern — HTTP request, assert response fields and values.

## Out of Scope

- **Product-level `discountType`**: only packages gain the field. The `products` table is untouched.
- **Frontend changes**: the API surface changes but UI adaptations are out of scope.
- **Historical checkout data**: existing `package_reserved` rows already captured the computed financial values at checkout time. No backfill or recalculation of past checkouts.

## Critical Implementation Notes

### Checkout math: three expressions must switch

The `InsertPackageReservedQuery` contains three separate expressions that currently hardcode `d.price * (1 - d.discount / 100)`:

1. **Tax** (`pr.tax`): `(p.platform_cut / 100.0) * d.price * (1.0 - d.discount / 100.0) * $3`
2. **Commission** (`pr.commission`): `(c.comission_value / 100.0) * d.price * (1.0 - d.discount / 100.0) * $3`
3. **Shop value** (`pr.shop_value`): `d.price * (1.0 - d.discount / 100.0) * $3`

**All three must be replaced** with `effective_unit_price(d.price, d.discount, d.discount_type) * $3`. Missing one will silently produce wrong financial numbers — no compiler error, no schema violation, just incorrect money in `package_reserved`.

After the change, each expression looks like:
- Tax: `(p.platform_cut / 100.0) * effective_unit_price(d.price, d.discount, d.discount_type) * $3`
- Commission (PERCENT): `(c.comission_value / 100.0) * effective_unit_price(d.price, d.discount, d.discount_type) * $3`
- Commission (FIXED): unchanged — uses `c.comission_value` directly
- Shop value: `effective_unit_price(d.price, d.discount, d.discount_type) * $3`

### Cross-package type chains

Two read paths span multiple packages. If a struct field is added in one package but the mapper in another package isn't updated, the field silently drops to zero/empty at the API boundary.

**Chain 1 — Consignment product listing (shop-facing):**

`ProductListInConsignmentQuery` → `ConsignmentProduct.Package` (consignmentmodel/data.go) → `PackageResource` and `ProductResource` (consignmentmodel/resources.go)

Fields to thread: `package_discount` → `Package.Discount`, `package_discount_type` → `Package.DiscountType`

**Chain 2 — Public consignment-by-code view (client-facing):**

`ProductsByConsignmentCode` → `ConsignmentProductRow` (productmodel/model.go) → `ProductConsignmentResource` (productmodel/resources.go → `ToConsignmentResource()`)

Fields to thread: `consignment_discount_type` → `ConsignmentDiscountType` → `DiscountType` on the resource

These chains span `consignmentmodel` and `productmodel` packages — a grep limited to one package will miss the other.

## Further Notes

- The migration is a single atomic DDL transaction: add column with DEFAULT, drop old CHECK, add new CHECK. Existing rows get `'percentage'` via DEFAULT.
- The `ProductListInConsignmentQuery` bug (returning product-level discount instead of package-level discount) is fixed in the same change because it's in the same query and the same data layer. It is not a separate issue.
- The `effective_unit_price` function uses `GREATEST(price - discount, 0)` to cap fixed discounts — a package cannot have a negative effective price.

Triage: ready-for-agent
