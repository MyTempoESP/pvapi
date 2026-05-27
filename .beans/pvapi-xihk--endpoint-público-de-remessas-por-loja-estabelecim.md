---
# pvapi-xihk
title: Endpoint público de remessas por loja (/estabelecimentos/:shop/remessas)
status: completed
type: feature
priority: normal
created_at: 2026-05-27T06:42:24Z
updated_at: 2026-05-27T06:54:12Z
parent: pvapi-sqf6
blocked_by:
    - pvapi-g798
---

## Problem Statement

Não existe endpoint público para listar remessas de uma loja específica.

## Solution

Criar `GET /estabelecimentos/:shop/remessas` — retorna remessas ativas de uma loja (filtradas por `consignments_active`).

## Implementation

- Nova rota no grupo shops: `public.GET("/estabelecimentos/:shop/remessas", shop.PublicListConsignmentsHandler(pool))`
- Novo handler `PublicListConsignmentsHandler` em `controllers/shop` — resolve shop por slug/ID, depois chama `consignmentmodel.List` (reusa função existente que já filtra por shopID e `consignments_active`)
- Nenhuma query ou model nova necessária
- Response: reusa `ConsignmentResource` existente

## Testing

Test de integração: seed uma remessa para a loja A, outra para a loja B. Chamar `/estabelecimentos/A/remessas` e verificar que só aparece a remessa da loja A.

## Summary of Changes

### Controller
- Added `PublicListConsignmentsHandler` in `controllers/shop/shop.go` — resolves shop by slug via `shopmodel.Resolve`, then calls `consignmentmodel.List` (reuses existing function that filters by shopID and `consignments_active`), returns response with `util.Wrap`.

### Routes
- Added `GET /estabelecimentos/:shop/remessas` in `routes.go` as a public route (no auth), following the same pattern as `GET /estabelecimentos/:shop/produtos`.

### Tests
- Added `public_list_by_shop_endpoint` subtest in `api_consignment_test.go` — creates two shops each with their own consignment, verifies each shop's endpoint returns only its own consignment and not the other's.
