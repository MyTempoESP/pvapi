---
# pvapi-6op6
title: Enriquecer listagem de clientes com total_purchases e last_purchase
status: completed
type: feature
priority: normal
created_at: 2026-06-01T19:54:05Z
updated_at: 2026-06-01T19:59:10Z
---

## Problem Statement

A shop owner looking at their client list sees only identity fields (id, name, email, phone). They cannot distinguish between a one-time buyer and a repeat customer, nor see when a client last purchased from them. This makes it impossible to prioritize high-value clients or track purchasing recency.

## Solution

Add two computed fields — total_purchases (sum of all shop_value) and last_purchase (most recent created_at) — to the client listing endpoint. These fields are derived from checkout_shop rows where client_id matches and shop_id equals the authenticated shop.

## User Stories

1. As a shop owner, I want to see each client's total spending at my shop, so that I can identify my most valuable customers.
2. As a shop owner, I want to see when each client last purchased from me, so that I can follow up with customers who haven't returned.
3. As a shop owner, I want the purchase data to reflect only completed checkouts attributed to a client, so that I'm not misled by carts that were never checked out.

## Implementation Decisions

- total_purchases is the SUM of checkout_shop.shop_value for a given client_id and shop_id. Serialized via decimal.Decimal (string in JSON).
- last_purchase is MAX(checkout_shop.created_at) for the same scope. Serialized as ISO 8601 time string via *time.Time with omitempty. NULL when the client has no checkouts.
- Both fields are added flat to ClientResource alongside existing id, name, email, phone.
- Query: a single SQL constant replacing ListShopClientsQuery. INNER JOIN on checkout_shop ensures only clients with actual purchases appear. No status filtering (all checkouts are 'paid').
- Defaults: decimal.Zero for total_purchases, nil for last_purchase.
- Struct shape (Client, adapted from prototype):

  TotalPurchases decimal.Decimal `db:"total_purchases"`
  LastPurchase   *time.Time      `db:"last_purchase"`

- Resource shape (ClientResource):

  TotalPurchases decimal.Decimal `json:"total_purchases"`
  LastPurchase   *time.Time      `json:"last_purchase,omitempty"`

- Clients without checkout_shop rows (anonymous checkouts, or never purchased) are excluded from the list.
- No new database function or migration needed — checkout_shop.client_id already exists (000036).

## Testing Decisions

- Test the new query in isolation against the test database to verify correct SUM/MAX aggregation and proper NULL handling.
- Follow existing patterns from api_checkout_client_test.go and client-related tests.
- A good test inserts checkout_shop rows with known values, calls the list endpoint, and asserts the aggregate fields.

## Out of Scope

- Filtering or sorting by total_purchases / last_purchase on the API.
- Including clients with zero purchases (empty state defaults shown as fallback).
- Any change to anonymous checkout attribution.
- Creating a reusable DB function for the aggregation.

## Further Notes

This change replaces the current cart-based client discovery (via cart_product → products → shop_id) with purchase-based discovery (via checkout_shop → shop_id). The existing endpoint /clientes already returns data scoped to the authenticated shop's ActiveRoleID, so no route or handler signature changes are needed.

Triage: ready-for-agent
