---
# pvapi-50dd
title: Adicionar campo retirada ao checkout confirm
status: todo
type: feature
priority: normal
created_at: 2026-06-18T16:45:26Z
updated_at: 2026-06-18T16:45:50Z
---

## Problem Statement

Clientes que compram produtos via checkout precisam escolher em qual Ponto (Pickup Location) vão retirar cada pacote. Atualmente o endpoint `/checkouts/confirm` não aceita um campo `retirada`, e o listing público de produtos não expõe o `id` dos Pontos disponíveis. Sem essa informação, o cliente não consegue indicar sua preferência de retirada, e o sistema não consegue registrar nem exibir o Ponto escolhido.

## Solution

Adicionar o campo `retirada` (ID do Ponto) ao request de confirmação de checkout, validar que o Ponto pertence à Remessa do pacote, armazená-lo em `package_reserved`, e devolvê-lo na resposta. Expor o `id` dos Pontos nos endpoints de listagem de produtos para que o frontend possa montar a seleção.

## User Stories

1. Como cliente, ao visualizar a listagem pública de uma Remessa, quero ver o `id` de cada Ponto disponível no array `retiradas`, para poder referenciá-lo ao confirmar a compra.
2. Como cliente, ao visualizar os detalhes de um produto (`/produtos/:id`), quero ver o `pickup_id` dentro de cada entrada do array `pickups`, para saber qual Ponto cada pacote pertence.
3. Como cliente, ao confirmar um checkout (`POST /checkouts/confirm`), quero poder enviar o campo `retirada` (ID do Ponto) em cada item do array `items`, para escolher onde retirar cada pacote.
4. Como cliente, se eu enviar um `retirada` que não pertence à Remessa daquele pacote, quero receber um erro `422` com uma mensagem clara, para corrigir minha seleção.
5. Como cliente, ao receber a resposta da confirmação do checkout, quero ver no array `qr_codes` o `id_retirada`, `nome_retirada` e `endereco` do Ponto escolhido, para confirmar visualmente onde devo retirar.
6. Como cliente, na resposta da confirmação, quero ver também `product_id`, `quantity`, `price` e `discount` em cada entrada de `qr_codes`, para ter um resumo do que comprei.
7. Como operador do sistema, quero que o `retirada` seja obrigatório no request de confirmação, para garantir que todo pacote reservado tenha um Ponto atribuído.
8. Como desenvolvedor, quero que a validação do `retirada` use a tabela `consignment_pickup_locations` como fonte da verdade, para garantir integridade referencial.
9. Como DBA, quero uma coluna `pickup_location_id NOT NULL` com FK em `package_reserved`, para que o banco imponha a integridade no nível do schema.

## Implementation Decisions

### Schema

- Nova coluna `pickup_location_id INTEGER NOT NULL` na tabela `package_reserved`, com foreign key para `pickup_locations(id)`.
- Migration up/down pair no diretório `internal/database/migrations/`.

### API: request

- `ConfirmItem` ganha campo `Retirada int` (JSON: `"retirada"`), obrigatório (`binding:"required"`).
- Validação: para cada item do checkout, verifica-se que o `retirada` está presente em `consignment_pickup_locations` para a Remessa do pacote. Em caso de falha, retorna HTTP 422 com mensagem `"retirada inválida para o pacote X"`.

### API: response de confirmação

- O array `qr_codes` deixa de ser `[{"package_id": N, "qr_code": "..."}]` e passa a ser:
  ```json
  {
    "qr_codes": [
      {
        "product_id": 1,
        "qr_code": "...",
        "quantity": 2,
        "price": "99.90",
        "discount": "0.00",
        "retirada": {
          "id": 5,
          "nome": "Shopping X",
          "address": { ... }
        }
      }
    ]
  }
  ```
- O objeto `retirada` dentro de cada `qr_code` contém `id`, `nome` e `address` (objeto de endereço completo, mesmo formato já usado em `PickupLocationResource`).

### API: listagem pública de produtos por Remessa

- O array `retiradas` (do endpoint público) ganha o campo `id` (int) em cada entrada. Mantém `name`, `address` como antes.
- O array `pickups` dentro de cada produto (no `PublicViewHandler`) ganha o campo `pickup_id` (int).

### Armazenamento

- A query `InsertPackageReservedQuery` é estendida para incluir `pickup_location_id`.
- O valor do `retirada` é propagado de `ConfirmItem.Retirada` → `writeCheckoutData` → `InsertPackageReservedQuery`.

### Módulos afetados

| Camada | O que muda |
|---|---|
| Migrations | Novo par up/down |
| Query layer | `PickupLocationsByConsignmentCodeQuery` inclui `id`; `InsertPackageReservedQuery` aceita `pickup_location_id`; nova query ou query inline para validar retirada × consignment |
| checkoutmodel | `ConfirmItem` ganha `Retirada`; `validateStock` ampliado ou nova `validateRetirada`; `writeCheckoutData` propaga o valor; `QRCodeEntry` vira struct enriquecida |
| productmodel | `PickupByCodeRow` ganha `PickupID` e mapeia no resource; `packageRow` e `BatchFetchPickups` propagam `PickupID` para `ProductPickupResource` |
| productmodel resources | `ConsignmentPickupResource` ganha `ID`; `ProductPickupResource` ganha `PickupID` |
| checkout controller | Tratamento de erro `422` para retirada inválida |

## Testing Decisions

- Testes de integração seguem o padrão existente em `api_checkout_client_test.go`: cenários de confirmação com e sem `retirada`, retirada inválida, e verificação do response enriquecido.
- Cada cenário cobre comportamento externo (status HTTP, corpo JSON), não detalhes de implementação.
- Adicionar ao menos: (1) checkout com retirada válida → confirma e retorna dados enriquecidos, (2) checkout com retirada inválida → 422, (3) checkout sem retirada → 400 (binding fail).

## Out of Scope

- Alterar endpoints de cancelamento, listagem de checkout por shop, client overview, ou consignments para expor retirada.
- Adicionar retirada ao carrinho ou a qualquer etapa anterior ao confirm.
- Migrar dados existentes (não há clientes reais ainda).
- Frontend.

## Further Notes

- A Remessa (`consignment`) é o escopo que define quais Pontos são válidos para um pacote, via `consignment_pickup_locations`.
- A validação no backend ignora `nome` e `address` vindos do request (se vierem); apenas o `id` é relevante.
- O campo `pickup_id` no `packages` view já existe no SQL que popula `ProductPickupResource` mas não era mapeado — é uma correção simples.

Triage: ready-for-agent
