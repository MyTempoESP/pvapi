---
# pvapi-m12y
title: Adicionar campo nome às remessas (consignments)
status: todo
type: feature
tags:
    - ready-for-agent
created_at: 2026-06-18T22:30:37Z
updated_at: 2026-06-18T22:30:37Z
---

## Problem Statement

Consignments (remessas) são identificadas atualmente apenas por um código UUID gerado automaticamente e um slug fixo. Donos de loja e monitores não têm como dar um rótulo legível às suas remessas. Nas listas, veem códigos como `a1b2c3d4-...` e slugs como `rem-generica`, inúteis no dia a dia.

## Solution

Adicionar um campo `name` obrigatório à entidade consignment. Campo de texto livre, visível em todas as respostas da API (listagem, detalhe, pública). Remessas existentes recebem backfill `"Sem nome"`.

## User Stories

1. Como dono de loja, quero nomear minhas remessas ao criá-las, para distinguir "Remessa Festival de Inverno 2026" de "Remessa Lote Promocional Outubro" no dashboard.
2. Como dono de loja, quero editar o nome de uma remessa existente, para corrigir erros de digitação ou atualizar o rótulo.
3. Como dono de loja, quero ver o nome da remessa nas listas, para encontrar a remessa certa sem memorizar UUIDs.
4. Como monitor (consignatário), quero ver o nome da remessa na minha lista, para identificar rapidamente em qual remessa estou trabalhando.
5. Como consumidor da API pública, quero ver o nome da remessa nas listagens públicas, para que usuários finais vejam rótulos significativos.

## Implementation Decisions

- **Nome do campo**: `name` (chave JSON), `Name` (campo Go). Consistente com a convenção camelCase em inglês dos DTOs de request (`destinationId`, `monitorProfitPercentage`).
- **Obrigatório na criação**: `binding:"required"` no DTO de request. Coluna no banco: `text NOT NULL`.
- **Backfill**: linhas existentes recebem `'Sem nome'` como default.
- **Editável no update**: request de update aceita campo `name` opcional (tipo ponteiro, nil = sem alteração).
- **Propagação via view**: coluna adicionada à view `consignments_listing_view`. Todas as queries de listagem, detalhe e pública leem dessa view — zero mudanças nas queries Go.
- **Tipo da coluna**: `text` sem limite. PostgreSQL trata `text` e `varchar` identicamente em armazenamento e performance; `text` evita um limite arbitrário.
- **Número da migration**: `000051` (sequencial, após `000050_add_pickup_location_id_to_package_reserved`).
- **`CreateConsignmentQuery`**: query SQL com named parameters aceita placeholder `@name`.
- **`ConsignmentBlueprint.Save`**: passa `Name` nos argumentos da query.
- **`ConsignmentUpdateBlueprint.Save`**: inclui condicionalmente `name` no mapa `SET` quando o ponteiro não é nil.
- **Structs de domínio afetadas**: `Consignment`, `ConsignmentWithStats`, `DetailedConsignment` (data.go); `ConsignmentBlueprint`, `ConsignmentUpdateBlueprint` (blueprint.go).
- **DTOs afetados**: `CreateConsignmentRequest`, `UpdateConsignmentRequest` (requests.go); `ConsignmentResource`, `ConsignmentWithStatsResource`, `PublicConsignmentResource` e seus métodos `ToResource`/`ToPublicResource` (resources.go).
- **Arquivos sem alteração**: controller (`controllers/consignment/consignment.go`), gate (`gate/consignment.go`), routes (`routes.go`), registro de policy (`controllers/policy/policy.go`). Camada de handlers e autorização não são afetadas — name é puramente dado.

## Testing Decisions

- Testes devem verificar que criar uma remessa com nome persiste corretamente e que listar/recuperar inclui o nome na resposta.
- Testes devem verificar que atualizar o nome de uma remessa funciona e que omitir o nome no update mantém o valor anterior.
- Testes devem verificar que remessas existentes (pré-migration) retornam `"Sem nome"`.
- Artefato de referência: testes de integração existentes em `api_consignment_test.go` cobrem CRUD de remessas; o mesmo padrão (chamadas HTTP autenticadas, asserções de JSON) se aplica.

## Out of Scope

- Restrições de unicidade no nome (nomes podem se repetir).
- Busca ou filtro por nome.
- Internacionalização da string de backfill.
- Mudanças de UI — escopo exclusivo de API.

## Further Notes

- A coluna `slug` atualmente hardcoded como `'rem-generica'` não é alterada por este PRD.
- A coluna `code` (UUID gerado automaticamente) também não é alterada.
