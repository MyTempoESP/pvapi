---
# pvapi-9679
title: Filtros por loja e remessa na listagem pública de produtos (/produtos)
status: completed
type: feature
priority: normal
created_at: 2026-05-27T06:42:35Z
updated_at: 2026-05-27T06:55:40Z
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

## Summary of Changes

### Implementation

- **`models/productmodel/model.go`**: Added `Consignments []int` field to `Product` struct (tagged `db:"-"` to avoid breaking existing scans). Modified `PublicIndex` to accept optional `shopIDs` and `consignmentIDs` parameters. Added internal `productRow` wrapper struct and `publicIndexWithFilters` function that builds dynamic SQL with squirrel for consignment filtering, including a LEFT JOIN subquery against `packages` + `consignments_active` views.

- **`controllers/product/product.go`**: Updated `PublicListingHandler` to parse `?shop=` and `?consignment=` repeatable query params from the request URL. Added `parseQueryInts` helper.

- **`models/productmodel/resources.go`**: Added `Consignments []int` field to `ProductResource` (JSON `consignments,omitempty`). Updated `ToResource()` to pass the consignment IDs through.

### How it works

- `?shop=1&shop=2` — filters to products from shops 1 and 2 (IN clause)
- `?consignment=5&consignment=6` — filters to products in active consignments 5 and 6; each product includes a `consignments` array of matching IDs
- Both combined: `?shop=1&consignment=5` — AND logic between the two filters

### Tests

Added three test cases to `TestProduct`:
- `filter_by_shop` — creates a second shop with its own product, verifies shop filter excludes other shops' products
- `filter_by_shop_multiple` — verifies repeatable shop param works
- `filter_by_consignment` — creates a consignment with a product, verifies consignment filter returns only products in those consignments with `consignments` array
