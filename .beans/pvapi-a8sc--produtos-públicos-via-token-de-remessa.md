---
# pvapi-a8sc
title: Produtos públicos via token de remessa
status: completed
type: feature
priority: normal
created_at: 2026-06-16T20:48:49Z
updated_at: 2026-06-16T21:37:46Z
---

# PRD: Produtos públicos via token de remessa

## Problem Statement

O endpoint público `/produtos` atualmente lista todos os produtos aprovados da plataforma, exigindo que clientes externos filtrem manualmente por loja (`?loja=`) ou remessa (`?remessa=`). Essa API expõe dados desnecessários (IDs internos, cortes de plataforma, status de moderação) e força o cliente a conhecer a estrutura interna para realizar checkouts. Clientes que possuem apenas o código da remessa (UUIDv4) não têm como obter os produtos específicos daquela remessa com os preços e estoques corretos (específicos do pacote).

## Solution

Modificar `/produtos` para exigir `?token=<código_da_remessa>` (UUIDv4). Sem token, retorna lista vazia. Com token válido, retorna os produtos da remessa com dados específicos do pacote (package_id como id, estoque/preço/desconto/destaque do pacote), remove campos internos (`platform_cut`, `platform_cut_type`, `moderation_status`) e inclui `retiradas` no nível raiz da resposta com nome e endereço dos pontos de retirada da remessa.

## User Stories

1. Como um cliente externo, quero obter os produtos de uma remessa usando apenas seu código UUID, para não precisar conhecer IDs internos.
2. Como um cliente externo, quero que o campo `id` do produto seja o `package_id`, para usar o mesmo valor diretamente no checkout sem tradução.
3. Como um cliente externo, quero ver estoque, preço, desconto e destaque específicos da remessa, para saber os valores corretos daquele contexto.
4. Como um cliente externo, quero que a resposta inclua apenas os pontos de retirada da remessa no nível raiz (`retiradas`), para saber onde retirar sem navegar produto a produto.
5. Como um cliente externo, quero receber lista vazia ao chamar `/produtos` sem token, para evitar exposição acidental de dados.
6. Como um cliente externo, quero receber erro 400 ao passar um UUID malformado, para saber que o token é inválido.
7. Como um cliente externo, quero receber erro 404 ao passar um UUID que não corresponde a nenhuma remessa, para diferenciar de outros erros.
8. Como um cliente externo, quero receber erro 410 com mensagem ao consultar uma remessa arquivada, para saber que a remessa não está mais ativa.
9. Como um desenvolvedor do frontend, quero que a resposta não inclua `platform_cut`, `platform_cut_type` e `moderation_status`, para não expor dados internos.
10. Como um desenvolvedor do frontend, quero receber `retiradas` como array de objetos com `name` e `address` (com street, number, neighborhood, city, state, cep, complement), para renderizar os locais de retirada.

## Implementation Decisions

1. **Busca de remessa por UUID**: nova função `consignmentmodel.FindByCode` que consulta a tabela `consignments` pelo campo `code`. Distingue entre remessa não encontrada e remessa arquivada (`status = 'ARQUIVADO'`).

2. **Novo tipo de recurso**: `ProductConsignmentResource`, derivado de `ProductResource`, com as seguintes alterações:
   - `id` mapeado de `package_id` (da tabela `packages`)
   - `stock` mapeado de `packages.quantity`
   - `price` mapeado de `packages.price`
   - `discount` mapeado de `packages.discount`
   - `featured` mapeado de `packages.featured`
   - Campos removidos: `platform_cut`, `platform_cut_type`, `moderation_status`, `pickups`

3. **Nova query SQL**: join entre `packages`, `product_catalog_view`, e `consignments` filtrado por `consignments.code = $1` e `product_moderation_status = 'approved'`, retornando campos do pacote e do produto.

4. **`retiradas` no nível raiz**: consulta separada via `consignment_pickup_locations` + `pickup_locations` + `aliased_addresses_view` para a remessa resolvida. Array de `{name, address}`.

5. **Controlador substituído**: `PublicListingHandler` trocado por nova função que:
   - Sem token → `{produtos: [], retiradas: []}`
   - Token malformado → 400
   - Token não encontrado → 404
   - Remessa arquivada → 410 com mensagem
   - Remessa ativa → produtos + retiradas

6. **Rota mantida**: `public.GET("/produtos", ...)` continua existindo, apenas troca o handler.

7. **Endpoint `/produtos/:product` mantido como está**: não documentado publicamente, sem alterações nesta PRD.

8. **Filtros `?loja=` e `?remessa=` removidos** do handler público de listagem.

## Testing Decisions

- **Módulos testados**: o handler substituído (`product.PublicListingHandler`) será coberto por testes de integração existentes em `api_product_test.go`.
- **O que testar**: comportamento external — resposta com/sem token, UUID inválido, remessa inexistente, remessa arquivada, formato da resposta (campos presentes/ausentes, `retiradas` no nível raiz).
- **Estilo de teste**: mesmo padrão do test suite existente — `api_product_test.go` usa `apiHarness` com `doRequest`, `decodeJSON`, `asMap`, `asSlice`, `asInt`, `asString` para verificar estrutura e valores.
- **Novos casos de teste**:
  - Listagem sem token retorna lista vazia
  - Listagem com token inválido retorna 400
  - Listagem com token inexistente retorna 404
  - Listagem com token de remessa arquivada retorna 410
  - Listagem com token válido retorna produtos com package_id como id
  - Produtos retornados têm estoque/preço/desconto/destaque do pacote
  - Resposta não contém `platform_cut`, `platform_cut_type`, `moderation_status`
  - Resposta contém `retiradas` no nível raiz com nome e endereço
  - Produto sem remessa não aparece na listagem por token (filtro approved)
  - `/produtos/:product` continua funcionando sem alterações
- **Testes existentes que serão afetados**: testes que chamam `/produtos` sem token esperando lista não-vazia — serão atualizados (produtos sem consignação não vão mais aparecer no endpoint público sem token).

## Out of Scope

- Remover `platform_cut*` e `moderation_status` do `/produtos/:product` (será feito em PRD futura)
- Alterar o endpoint `/estabelecimentos/:shop/produtos`
- Alterar endpoints autenticados de produto
- Remover as colunas `platform_cut` e `platform_cut_type` do banco de dados
- Migração de schema

## Further Notes

- O campo `featured` já existe na tabela `packages` e é populado pelo `ProductAddToConsignmentQuery`.
- A view `consignments_active` exclui remessas com status `ARQUIVADO` ou `RASCUNHO`, mas a resolução por UUID busca diretamente na tabela `consignments` para poder distinguir entre remessa ativa e arquivada.
- O código da remessa (`code`) é gerado como `uuidv4()` no `CreateConsignmentQuery`.

Triage: ready-for-agent



## Summary of Changes

### Problem
O endpoint público  listava todos os produtos aprovados sem restrição, expondo dados internos (platform_cut, moderation_status) e forçando clientes a conhecer IDs internos para filtrar.

### Solution
Modificado  para exigir  (UUIDv4):

**Handler ():**
- Sem token → retorna 
- Token malformado → 400
- Token não encontrado → 404
- Remessa arquivada → 410
- Remessa ativa → produtos + retiradas

**Novo recurso ():**
-  — sem , , , 
- uid=1001(rodrigo) gid=1001(rodrigo) grupos=1001(rodrigo),20(dialout),27(sudo),100(users),1000(isabela),1002(sims) mapeado de ; , , ,  do pacote
-  com  no nível raiz

**Consulta por código ():**
-  — join packages + product_catalog_view filtrado por consignments.code

**Lookup de remessa ():**
-  — busca na tabela  por code, retorna ID + status para distinguir ativo/arquivado

**Removido:** antigos filtros  e  do handler público, função  do controller de produto

**Testes:** 17 testes em TestProduct, 30 em TestConsignment — todos passam
