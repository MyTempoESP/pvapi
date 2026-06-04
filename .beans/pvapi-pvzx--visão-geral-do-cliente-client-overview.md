---
# pvapi-pvzx
title: Visão geral do cliente (Client Overview)
status: done
type: feature
priority: normal
created_at: 2026-06-04T22:44:08Z
updated_at: 2026-06-04T22:44:13Z
---

## Problem Statement

A client who has completed one or more checkouts on the platform has no way to see a consolidated summary of their purchase history. They must rely on individual confirmation emails and have no in-app view of what they bought, what's pending, or where to pick up their packages.

## Solution

Add a `GET /usuario/cliente/overview` endpoint that returns a client's purchase summary: total checkout count, pending package count, and a list of packages with pickup location and product details. The client authenticates via their JWT (which carries their client ID as the ActiveRoleID for pure-client users).

Response shape:

```
{
  "success": true,
  "message": "visão geral carregada",
  "orders": 5,
  "pending": 2,
  "packages": [
    {
      "uuid": "<checkout.unique_id>",
      "date": "<package_reserved.reserved_at>",
      "status": "<package_reserved.status>",
      "value": "<package_reserved.shop_value>",
      "quantity": "<package_reserved.quantity>",
      "pickup": {
        "id": "<pickup_location.id>",
        "name": "<pickup_location.name>",
        "datetime": "<pickup_location.pickup_at>",
        "address": {
          "street": "...",
          "number": ...,
          "neighborhood": "...",
          "city": "...",
          "state": "...",
          "cep": "...",
          "complement": "..."
        }
      },
      "product": {
        "id": "<product.id>",
        "name": "<product.name>",
        "code": "<product.code>"
      }
    }
  ]
}
```

## User Stories

1. As a client, I want to see how many checkouts I've completed, so that I have a quick summary of my purchase activity.
2. As a client, I want to know how many packages are still pending (not yet received), so that I can plan my pickups.
3. As a client, I want to see a list of all my packages with their current status, so that I can track each item's progress.
4. As a client, I want to know the pickup location (name, date, address) for each package, so that I know where and when to collect them.
5. As a client, I want to know which product each package contains (name and code), so that I can identify what I bought.
6. As a client, I want to see the value I paid per package, so that I can verify charges.
7. As a client, I want to see the checkout unique ID (UUID) for each package, so that I can reference it in support communications.
8. As a client, I want my packages ordered by most recent first, so that my newest purchases appear at the top.

## Implementation Decisions

### Authentication

The endpoint authenticates via the existing JWT. For pure-client users (users who exist only in the clients table, with no admin, shop, or consignee role), the JWT's ActiveRoleID claim already carries the clients.id. Users with multiple roles (e.g., shop owner who is also a client) cannot use this endpoint to view their client purchases — their ActiveRoleID points to their shop. This is acceptable because a single email cannot have two accounts on the platform.

### Data sources

Client packages are linked through the join chain: package_reserved -> checkouts (via checkout_id) -> carts (via cart_id) -> client_id. This means only checkouts where a client_id was set on the cart (checkouts with ClientEmail provided) are included. Anonymous checkouts are excluded.

### Queries

Two separate read queries, following the shopmodel.BuildCheckoutReport pattern:

1. **Counts query**: aggregates COUNT(DISTINCT checkout_id) for total checkouts and COUNT(*) WHERE status != 'recebido' for pending packages, filtered by the client's ID via the join chain.
2. **Packages query**: returns package_reserved rows joined through checkouts -> carts -> packages -> products -> consignments -> pickup_locations -> addresses, ordered by reserved_at DESC.

No pagination. The overview returns all packages in a single response.

### Route

`GET /usuario/cliente/overview` — placed under a new route group on `/usuario/cliente` with auth middleware. The `/usuario` prefix is added to the default auth switch in the policy bouncer, allowing any authenticated user (no role gate required). The handler internally resolves the client ID and returns 404 if the user is not a client.

### Response conventions

Follows existing patterns: wrapped response with success, message, and the three data fields (orders, pending, packages). Decimal values use decimal.Decimal for JSON serialization, consistent with other endpoints.

### Statuses

package_reserved.status has two possible values: 'pending' and 'recebido'. The pending count uses `status != 'recebido'`. No additional status values exist at this time.

### No product images

Product images (stored in the attr JSONB column) are not included in the overview response. Only id, name, and code are returned.

## Testing Decisions

Tests should follow the existing integration test pattern used by api_*_test.go files — test against a real temporary PostgreSQL database, exercise the full HTTP handler through the Gin router, and assert on JSON response shape and correctness.

- Test overview for a client with no checkouts (zero counts, empty packages array).
- Test overview for a client with one checkout containing multiple packages (verify counts, package list).
- Test overview for a client with a mix of pending and received packages (verify pending count).
- Test overview for a non-client user (expect 404).
- Test overview for an unauthenticated request (expect 401).

Good tests verify external behavior: HTTP status codes, JSON field presence and types, counts match database state. Do not test internal query strings or model structs directly.

Prior art: api_checkout_client_test.go, api_cart_checkout_test.go, api_client_test.go.

## Out of Scope

- Pagination of the packages array.
- Product images in the overview response.
- Client role extension to the JWT system (shop+client dual-role users cannot use this endpoint).
- The ability for a client to mark packages as received (separate endpoint).
- Filtering or searching the packages list.
- WebSocket or real-time updates.

## Further Notes

The domain glossary defines "Checkout" as the canonical term (avoid "Order"). The API response field "orders" intentionally uses the colloquial term because it's the user-facing contract — clients think in terms of "my orders", not "my checkouts". This is a deliberate UX choice, not a glossary violation in the data layer.

Triage: ready-for-agent
