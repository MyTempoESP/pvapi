---
# pvapi-sqf6
title: 'API pública: expor remessas e enriquecer listagem de produtos'
status: todo
type: epic
created_at: 2026-05-27T06:42:07Z
updated_at: 2026-05-27T06:42:07Z
---

## Problem Statement

O frontend público (loja online) precisa consumir dados de remessas e filtrar produtos por loja e remessa. Atualmente:

- A listagem pública de produtos (`/produtos`) retorna todos os produtos aprovados, sem filtro por loja ou remessa
- Não existe endpoint público para listar remessas — todas as rotas `/remessas` exigem autenticação
- O endpoint por loja (`/estabelecimentos/:shop/remessas`) também não existe

## Solution

Três endpoints públicos:

1. **`GET /remessas-publicas`** — listagem pública de todas as remessas ativas (excluindo RASCUNHO e ARQUIVADO)
2. **`GET /estabelecimentos/:shop/remessas`** — listagem pública de remessas de uma loja específica (excluindo RASCUNHO e ARQUIVADO)
3. **`GET /produtos?shop=1&shop=2&consignment=5&consignment=6`** — listagem pública de produtos com filtros opcionais por loja (repeatable, AND) e por remessa (repeatable, AND, apenas consignments_active)

## User Stories

1. As a site visitor, I want to see all active consignments, so that I can browse what's available across all shops
2. As a site visitor browsing a specific shop, I want to see that shop's active consignments, so that I can find products organized by remessa
3. As a site visitor, I want to filter products by one or more shops, so that I can compare products across specific lojas
4. As a site visitor, I want to filter products by one or more consignments, so that I can see all products within specific remessas
5. As a site visitor combining both filters, I want to see only products that belong to BOTH the selected shops AND the selected consignments (AND logic), so that I can narrow my search precisely
6. As a site visitor filtering by consignment, I want each product to indicate which consignments it belongs to, so that I understand the product's context

## Implementation Decisions

- Route `/remessas-publicas` uses a separate path (not overriding `/remessas`) to avoid middleware conflicts with the authenticated group
- Route `/estabelecimentos/:shop/remessas` follows the existing pattern of `/estabelecimentos/:shop/produtos`
- Both consignment endpoints use `consignments_active` view (excludes RASCUNHO, ARQUIVADO)
- The per-shop consignment endpoint reuses the existing `consignmentmodel.List` function, which already takes shopID and filters via `consignments_active`
- The product listing filter builds dynamic SQL with squirrel:
  - `?shop=` adds `WHERE product_shop_id IN (...)`
  - `?consignment=` adds `LEFT JOIN packages` + `WHERE consignment_id IN (...)` gated by `consignments_active`
  - Both filters are AND-combined when present together
- When `?consignment` is present, each product in the response includes a `consignments` array with consignment IDs. When absent, no consignment metadata is returned
- Response types reuse existing resources (`ConsignmentResource`, `ProductResource`) with minimal extensions

## Testing Decisions

- Test #1: Seed a consignment, hit `/remessas-publicas`, verify it appears in the response
- Test #2: Seed a consignment for a specific shop, hit `/estabelecimentos/:shop/remessas`, verify it appears and consignments from other shops do not
- Test #3: Test both filters on `/produtos` — verify `?shop=` filtering and `?consignment=` filtering work correctly (no edge cases needed)
- Follow existing test patterns: seed data via DB, call endpoint, assert response shape and content
- Tests should use the project's test-code skill (temporary PostgreSQL database)

## Out of Scope

- Pagination on any of the new endpoints
- Sorting/filtering by consignment status beyond the active/draft/archived split
- Public product listing enriched with consignment data when consignment filter is NOT active
- Any changes to authenticated consignment/product endpoints

## Further Notes

- Dependency: Bean #2 (`/estabelecimentos/:shop/remessas`) is blocked by Bean #1 (`/remessas-publicas`)
- Bean #3 is independent — it only needs the `consignments_active` view and packages JOIN, both of which already exist in the schema
- The `consignments_active` view already exists in the database (migration 000025)
