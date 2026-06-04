---
# pvapi-85io
title: Add pickup locations to /produtos and /produtos/:id routes
status: completed
type: feature
priority: normal
created_at: 2026-06-04T03:29:15Z
updated_at: 2026-06-04T04:02:06Z
---

## Problem Statement

A client browsing products on the marketplace cannot see where a product is available for pickup without going through the cart flow. The cart's `/carrinho/pacotes` endpoint provides package-level pickup options, but it requires a cart session — adding friction and making the cart an unnecessary intermediary.

## Solution

Embed pickup location availability directly into the `/produtos` (public listing) and `/produtos/:product` (public view) API responses. Each product gains a `pickups` array containing every available package option — with its price, discount, quantity, pickup location name, date, address, and package ID. This eliminates the separate cart-based package lookup and moves the decision point earlier in the browsing flow.

## User Stories

1. As a marketplace customer, I want to see available pickup locations for a product on the product listing page, so that I can compare pickup options before adding anything to a cart.
2. As a marketplace customer, I want to see available pickup locations on the product detail page, so that I know where and when I can collect the product before purchasing.
3. As a marketplace customer, I want to see the discounted price per package option, so that I can make informed decisions about which pickup to choose.
4. As a marketplace customer, I want to see the stock quantity per package option, so that I know if the product is still available at that pickup.
5. As a marketplace customer, I want to see the pickup date and location address, so that I can choose a pickup that fits my schedule and proximity.
6. As a frontend developer, I want a consistent `pickups` array shape across listing and detail endpoints, so that I can reuse the same rendering component.
7. As a frontend developer, I want the `pickups` array to be empty rather than null when no packages exist, so that I avoid null-check boilerplate.
8. As a user filtering products by consignment, I still want to see the pickups data for the filtered products, so that the filter remains useful while surfacing richer data.
9. As a user filtering products by shop, I still want to see the pickups data, so that shop-filtered browsing benefits from the same pickup visibility.

## Implementation Decisions

### API Contract

Each product in the `/produtos` and `/produtos/:product` responses gains a `pickups` field:

```json
{
  "id": 1,
  "name": "Medalha Personalizada",
  "price": "19.90",
  "discount": "2.00",
  "stock": 15,
  "...": "...",
  "pickups": [
    {
      "package_id": 42,
      "price": "3.00",
      "discount": "0.50",
      "quantity": 10,
      "pickup_name": "Feira da Sé",
      "pickup_date": "2026-06-15T14:00:00",
      "address": {
        "street": "Rua X",
        "number": "123",
        "neighborhood": "Centro",
        "city": "São Paulo",
        "state": "SP",
        "cep": "01001-000",
        "complement": "nenhum"
      }
    }
  ]
}
```

- `package_id` is `packages.id` — the frontend uses this to confirm checkout.
- `price` and `discount` are separate fields (not merged), matching the `packages` table columns.
- `quantity` is `packages.quantity`.
- `pickup_name` and `pickup_date` are included to avoid a follow-up query for display metadata.
- When a product has no available packages, `pickups` is `[]` (empty array), never `null`.

### Availability Filters

Only packages matching these conditions are included:
- The consignment is active (`status NOT IN ('ARQUIVADO', 'RASCUNHO')`)
- `packages.quantity > 0`
- The product itself is `moderation_status = 'approved'` (already enforced by the product queries)

These are identical to the filters currently used by the cart's `/carrinho/pacotes` endpoint.

### Removed: Consignments Field

The existing `consignments` field (array of consignment IDs) on `ProductResource` is removed. The `pickups` array supersedes it — every pickup entry implicitly belongs to a consignment, and the richer shape (price, quantity, address) makes the flat ID list redundant.

### Removed: Cart Packages Endpoint

The `/carrinho/pacotes` route, its handler (`PackageOptionsHandler`), model code (`FindPackageOptions`, `findAvailablePackages`, `findCartProducts`, `groupByProduct`), queries (`AvailablePackages`, `CartProductsForPackages`), and types (`PackageOption`, `PackageOptionResource`, `ProductPackageOptions`, `CartPackageProduct`) are all deleted.

### Modules

**New module — product pickup aggregation**: A batch query fetches all packages for a set of product IDs in a single round-trip. Results are grouped by product ID in application code (two-query pattern: products query + pickups query, merge in Go). This avoids N+1 queries on product listings.

**Modified module — product public index**: The consignment-filter path (`publicIndexWithFilters`) drops the `product_consignments` subquery column and the `productRow` helper. The `EXISTS` WHERE clause for consignment filtering remains.

**Modified module — product resource**: `ProductResource` gains `Pickups` and loses `Consignments`.

**Deletion module — cart packages**: The entire `models/cartmodel/package.go` file, the `AvailablePackages` and `CartProductsForPackages` queries, and the `PackageOptionsHandler` controller method.

### Endpoint Scope

Only two public endpoints include pickups:
- `GET /produtos` — product listing (with optional `?shop=` and `?consignment=` filters)
- `GET /produtos/:product` — product detail view

The shop-product listing (`/estabelecimentos/:shop/produtos`) does *not* include pickups. Because `ProductResource.Pickups` is populated only in the two target handlers (not by `PublicIndexByShop`), shop-listed products exclude them.

### Two-Query Aggregation Strategy

For the listing endpoint:
1. Execute the existing product query (with shop/consignment filters)
2. Collect the resulting product IDs
3. If any IDs exist, execute a single batch query joining `packages → consignments_active → pickup_locations → aliased_addresses_view` with `packages.product_id IN (...)`, `quantity > 0`
4. Group package rows by `product_id` in Go
5. Attach grouped pickups to each product

For the single-product endpoint, the same batch query pattern applies (with a single product ID).

## Testing Decisions

### What Makes a Good Test

Tests verify the external HTTP contract — the shape and presence of the `pickups` field in API responses — not the internal query implementation. The existing integration test harness (`newAPIHarness`) is used for all test cases.

### Test Cases

Tests are added to the existing `api_product_test.go` (the `TestProduct` function), following prior art for how fixtures and assertions are structured in that file:

1. **Pickups on product detail**: View a product via `GET /produtos/:id` after adding it to a consignment. Assert `pickups` is present, non-empty, and each entry contains the expected shape (`package_id`, `price`, `discount`, `quantity`, `pickup_name`, `pickup_date`, `address`).
2. **Pickups on product listing**: List all products via `GET /produtos`. Assert products that belong to active consignments have non-empty `pickups`.
3. **Empty pickups on product without consignments**: View a product that has never been added to a consignment. Assert `pickups` is `[]`.
4. **No pickups on shop-listed products**: Call `/estabelecimentos/:shop/produtos`. Assert the products in the response do NOT contain a `pickups` field.
5. **Consignments field removed**: View a product. Assert the response does NOT contain a `consignments` field.

### Prior Art

The existing `TestProduct` function tests the same endpoints (`GET /produtos`, `GET /produtos/:id`) and uses the same fixture helpers (`createProduct`, `createConsignment`, `addProductToConsignment`, `approveProduct`). New tests follow the same patterns: create fixtures, hit the endpoint, decode JSON, assert on shape.

## Out of Scope

- Cart removal (carts are being removed separately)
- Any changes to other product endpoints (`/estabelecimentos/:shop/produtos`, `/remessas/:consignment/produtos`, shop-authenticated product listing)
- Database schema changes (none needed — `packages`, `consignments`, `pickup_locations` already have the data)
- Checkout flow changes (checkout already accepts `package_id` and is unaffected)

## Further Notes

- The checkout endpoint already references `package_id` via `ConfirmItem.PackageID`, so the frontend can take the `package_id` from the product's `pickups` array and use it directly in checkout confirmation.
- The `getPackageID` test helper in `api_fixtures_test.go` uses `/carrinho/pacotes` to discover package IDs. This helper will need updating or removal as part of the cart removal workstream.



Triage: ready-for-agent



## Summary of Changes

### Added
- **Query**  in  — batch-fetches packages (with pickup location, address) for a set of product IDs
- **Types** ,  in  — pickup data shape per the API contract
- **Function**  in  — takes product IDs, queries packages, groups by product ID, initializes empty slices for products without packages
- **Pickups field** on  — uses  with  so it's absent on shop-listed products but  on public endpoints when empty
- **Tests** in  — , , , , 

### Modified
- **** and **** — enrich products with pickups via  before serializing
- **** test helper — uses  pickups array instead of removed 
- **** — removed  struct and  subquery/column, keeping only the  WHERE clause for consignment filtering
- **** — shop2 package ID lookup uses product detail endpoint instead of removed cart packages endpoint
- **** replaces  in the model struct

### Removed
- **** — entire file (types: , , , , ; functions: , , , )
- **** from 
- **** route from 
- **** and **** queries from 
- ** field** from  and  struct
- **Stale test cases** in  — , , , 
- ** field** from  model struct
