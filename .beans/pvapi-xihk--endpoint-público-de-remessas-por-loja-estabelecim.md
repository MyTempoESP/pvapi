---
# pvapi-xihk
title: Endpoint público de remessas por loja (/estabelecimentos/:shop/remessas)
status: todo
type: feature
created_at: 2026-05-27T06:42:24Z
updated_at: 2026-05-27T06:42:24Z
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
