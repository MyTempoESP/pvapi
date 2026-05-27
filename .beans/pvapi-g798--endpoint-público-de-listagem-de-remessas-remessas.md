---
# pvapi-g798
title: Endpoint público de listagem de remessas (/remessas-publicas)
status: completed
type: feature
priority: normal
created_at: 2026-05-27T06:42:15Z
updated_at: 2026-05-27T06:48:20Z
parent: pvapi-sqf6
---

## Problem Statement

Não existe endpoint público para listar remessas. Todas as rotas `/remessas` estão sob autenticação.

## Solution

Criar `GET /remessas-publicas` — retorna todas as remessas ativas (filtradas por `consignments_active`, excluindo RASCUNHO e ARQUIVADO).

## Implementation

- Nova rota: `public.GET("/remessas-publicas", consignment.PublicListHandler(pool))`
- Novo handler `PublicListHandler` em `controllers/consignment`
- Nova função `consignmentmodel.PublicList` — igual ao `ListConsignmentsByShopQuery` mas sem filtro de shop_id
- Response: reusa `ConsignmentResource` existente
- Nova query SQL: `ListAllActiveConsignmentsQuery`

## Testing

Test de integração: seed uma remessa ativa, chamar o endpoint, verificar que aparece na resposta.
