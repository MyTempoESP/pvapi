---
# pvapi-9679
title: Filtros por loja e remessa na listagem pública de produtos (/produtos)
status: todo
type: feature
created_at: 2026-05-27T06:42:35Z
updated_at: 2026-05-27T06:42:35Z
parent: pvapi-sqf6
---

## Problem Statement

A listagem pública de produtos (`/produtos`) retorna todos os produtos aprovados sem possibilidade de filtrar por loja ou remessa.

## Solution

Adicionar query strings `?shop=` e `?consignment=` ao endpoint `GET /produtos`:

- `?shop=1&shop=2` — filtra produtos das lojas especificadas (AND entre filtros, OR dentro do mesmo filtro)
- `?consignment=5&consignment=6` — filtra produtos que pertencem às remessas especificadas (AND com shop, apenas `consignments_active`)
- Ambos combinados: produtos que estão nas lojas E nas remessas (AND lógico)

## Implementation

- Modificar `product.PublicListingHandler` para parsear `?shop` (repeatable ints) e `?consignment` (repeatable ints) da query string
- Modificar `productmodel.PublicIndex` para aceitar slices opcionais de shop IDs e consignment IDs
- SQL dinâmico com squirrel:
  - Sempre: `FROM product_catalog_view WHERE product_moderation_status = 'approved'`
  - Se `?shop`: adiciona `AND product_shop_id IN (...)`
  - Se `?consignment`: adiciona `LEFT JOIN packages ON ...` + `AND consignment_id IN (SELECT id FROM consignments_active WHERE id IN (...))`
- Quando filtro `?consignment` está presente: cada produto inclui campo `consignments` (array de IDs). Quando ausente, não inclui

## Testing

Test de integração: seed produtos em lojas diferentes, alguns associados a remessas. Verificar:
- `?shop=X` retorna apenas produtos da loja X
- `?consignment=Y` retorna apenas produtos na remessa Y
