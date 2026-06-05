---
# pvapi-n918
title: 'Shop sales endpoint: GET /clients/:id/sales'
status: todo
type: feature
priority: normal
created_at: 2026-06-05T06:34:02Z
updated_at: 2026-06-05T06:34:06Z
---

## Problem Statement

Shop operators need to view a client's purchase history to understand what products the client has bought, how much they've spent, and the status of each purchase. Currently there is no endpoint that exposes this per-client sales data.

## Solution

A new authenticated endpoint `GET /clients/:id/sales` accessible by shop users. Returns a flat JSON array of sales belonging to the specified client, filtered to the authenticated shop. Each sale includes its ID, client ID, product names with quantities, total value, date, and aggregated status.

**Response shape:**

```json
[
  {
    "id": "42-5",
    "clientId": 5,
    "products": ["2x Product A", "1x Product B"],
    "total": 15.50,
    "date": "2026-06-05T00:00:00.000Z",
    "status": "Concluída"
  }
]
```

## User Stories

1. As a shop operator, I want to list all sales for a specific client, so that I can review their purchase history.
2. As a shop operator, I want to see which products a client bought in each sale, so that I understand their preferences.
3. As a shop operator, I want to see the total value of each sale, so that I can track the client's spending.
4. As a shop operator, I want to see when each sale occurred, so that I can understand purchase recency.
5. As a shop operator, I want to see whether a sale is completed or still pending, so that I know which orders have been fully received.
6. As a shop operator, I want each sale to be scoped to my shop only, so that I don't see purchases from other shops.
7. As a shop operator, I want to see product quantities in the product list, so that I can distinguish between sales with the same products but different amounts.
8. As an unauthenticated user, I should be denied access to the sales endpoint.
9. As a non-shop user, I should be denied access to the sales endpoint.
10. As a shop operator, I should receive a 404 when querying sales for a non-existent client.
11. As a shop operator, I should receive an empty array when the client has no sales with my shop.

## Implementation Decisions

**Data model:** One sale maps to one `checkout_shop` row — the junction between a checkout and a shop that already scopes purchases by shop and carries shop_value. Products are sourced from associated `package_reserved` rows.

**Sale ID:** Composite string `"checkout_id-shop_id"` since `checkout_shop` has a composite primary key of `(checkout_id, shop_id)`.

**Total:** `checkout_shop.shop_value` — the gross amount the shop receives before platform tax/commission.

**Products:** Array of strings with quantities, e.g. `"2x Product A"`. Derived from `package_reserved` quantity × product name. Grouped per checkout_shop.

**Status:** Aggregated from `package_reserved.status` across all packages in the sale. All `'recebido'` → `"Concluída"`. Any `'pending'` → `"Pendente"`.

**Date:** `checkout_shop.created_at` — the sale row's own timestamp.

**Routing:** New route group `/clients` (distinct from existing `/clientes`), registered with auth middleware. Single route: `GET :id/sales`.

**Authorization:** A new policy guard registered at `/clients` reusing the existing `ClientGate`. The gate dispatches `GET :id/sales` to a new `Sales()` method that only allows `role == "shop"`.

**Response format:** Bare top-level JSON array (matching the frontend contract).

**Query strategy:** SQL fetches flat rows joined across `checkout_shop` → `package_reserved` → `packages` → `products`. Grouping, product name composition, and status derivation are done in Go (consistent with how ClientOverview handles package rows).

**Client validation:** The endpoint verifies the client exists (via a lightweight existence check on the clients table). Returns 404 for non-existent or non-client users.

**No schema migrations required.** All data already exists in `checkout_shop`, `package_reserved`, `packages`, and `products`.

## Testing Decisions

**What makes a good test:** The endpoint is a read-only HTTP GET. Tests should verify HTTP status codes, response shape, and field semantics against a real database. Internal SQL or grouping logic is tested indirectly via assertions on the JSON output.

**Integration test file** covering:
- Unauthenticated request → 401
- Non-shop user request → 403
- Client does not exist → 404
- Client exists but has no sales → 200 with empty array
- Client with one checkout → 200 with correct sale fields (id, clientId, products, total, date, status)
- Client with multiple checkouts → 200 with multiple sales ordered by date
- Client with both pending and received packages → correct aggregated status

**Pattern:** Follows existing integration test conventions from `api_client_test.go` and `api_client_overview_test.go` — uses the shared `apiHarness`, creates fixtures through the API, and calls the new endpoint with `doRequest`.

## Out of Scope

- Pagination
- Filtering by date range or status
- Sorting beyond default date-descending
- Sales from a different shop's perspective (the endpoint is always scoped to the authenticated shop)
- Anonymous/null-client sales (these have no client_id, so are invisible by definition)
- Mutation of sales (update, cancel, refund)

## Further Notes

The URL `/clients/:id/sales` uses English naming (`/clients` instead of `/clientes`) to match the frontend contract, even though the rest of the API uses Portuguese route prefixes. The existing `/clientes` group remains unchanged.

**Triage: ready-for-agent**

Triage: ready-for-agent
