---
# pvapi-5t7q
title: Remover produto de remessa (DELETE /remessas/:id/produtos/:product)
status: completed
type: feature
priority: normal
created_at: 2026-05-27T15:49:56Z
updated_at: 2026-05-27T16:31:22Z
---

## Problem Statement

Lojas não conseguem remover um produto individual de uma remessa (consignment) após adicioná-lo. A única operação destrutiva disponível é arquivar a remessa inteira, o que é drástico demais para correções pontuais (ex.: produto adicionado por engano, mudança de sortimento).

Atualmente a API expõe:
- `POST /remessas/:consignment/produtos` — adiciona produtos
- `GET /remessas/:consignment/produtos` — lista produtos
- `PUT /remessas/:consignment/produtos/:product` — atualiza preço/quantidade

Falta o `DELETE` correspondente.

## Solution

Adicionar um endpoint `DELETE /remessas/:consignment/produtos/:product` que:

1. Valida que a remessa não está arquivada (status `ARQUIVADO`)
2. Valida que não há reservas ativas em `package_reserved` para o produto na remessa
3. Lê a quantidade atual do produto na remessa
4. Incrementa `products.stock` com essa quantidade
5. Remove a linha de `packages` para o par (product_id, consignment_id)

Tudo dentro de uma única transação PostgreSQL. Respeita o mesmo controle de acesso do `PUT` (apenas a loja dona da remessa).

## User Stories

1. Como lojista, quero remover um produto específico de uma remessa, para corrigir erros de sortimento sem precisar recriar a remessa inteira.
2. Como lojista, quero que o estoque do produto seja restaurado automaticamente ao removê-lo da remessa, para não perder inventário.
3. Como lojista, quero receber um erro claro se tentar remover um produto que não está na remessa (404), para entender o que deu errado.
4. Como lojista, quero que apenas minha própria loja possa remover produtos das minhas remessas, para evitar manipulação por terceiros.
5. Como consignatário (monitor), NÃO quero poder remover produtos de remessas — apenas visualizá-los — mantendo a separação de responsabilidades atual.
6. Como lojista, quero que a remoção seja atômica (restaura estoque E remove o registro na mesma transação), para evitar inconsistências.
7. Como lojista, quero receber um erro claro (400) se tentar remover um produto de uma remessa já arquivada, para não corromper dados históricos.
8. Como lojista, quero receber um erro claro (409) se o produto tiver reservas ativas em checkouts pendentes, para não corromper vendas em andamento.

## Implementation Decisions

### Novo endpoint

`DELETE /remessas/:consignment/produtos/:product` — recebe os dois IDs pela URL, sem body. Retorna `200` com mensagem "produto removido da remessa com sucesso".

### Módulo: `product_increment_inventory` (migração 000033)

Nova função PL/pgSQL simétrica a `product_decrement_inventory`. Mantém o `FOR UPDATE` row lock para consistência, mas sem validação de estoque (incrementar nunca falha por "estoque insuficiente").

### Módulo: `RemoveProductAction`

Struct com método `Exec(ctx, tx)` que encapsula toda a lógica transacional:

1. Verificar que a remessa não está `ARQUIVADA` — se estiver, retornar 400
2. Verificar que não há reservas ativas em `package_reserved` para o pacote — se houver, retornar 409
3. Ler a quantidade atual do pacote (reusa `VerifyPackageQuantity`)
4. Chamar `product_increment_inventory` para devolver ao estoque
5. Executar `DELETE FROM packages WHERE product_id = $1 AND consignment_id = $2`

A operação é envolvida por `util.TransactionAware` no controller, herdando o padrão de `AddProductHandler`.

### Controle de acesso

Adicionar `r.Method == "DELETE"` retornando `c.Modify(u, r)` no case `/produtos/:product` do `ConsignmentGate.Dispatch`. O método `Modify` chama `Owns`, que verifica `shopOwns` — apenas a loja dona da remessa pode remover produtos.

### Erros

| Cenário | Status | Mensagem |
|---|---|---|
| Produto não está na remessa | 404 | "produto não encontrado na remessa" |
| Remessa arquivada | 400 | "não é possível remover produtos de uma remessa arquivada" |
| Reservas ativas (package_reserved) | 409 | "não é possível remover: produto possui reservas ativas em checkouts" |
| Sucesso | 200 | "produto removido da remessa com sucesso" |
| Acesso negado (gate) | 403 | tratado pelo middleware de autorização |

### Bloqueio de produtos duplicados

Criado o bean `pvapi-dnjn` para bloquear a adição de produtos duplicados na mesma remessa. Isso garante que o DELETE sempre opera sobre uma única linha em `packages` para o par (product_id, consignment_id), simplificando a lógica de remoção.

### Reservas de checkouts recebidos

Reservas de checkouts com status "received" devem ser removidas no momento da confirmação de recebimento (anotado como lembrete para o time). Até lá, qualquer reserva em `package_reserved` bloqueia a remoção com 409.

## Testing Decisions

### O que testar

Seguir o padrão de `api_consignment_test.go`: harness-based, subtests com `t.Run`, verificação de status code e body JSON. Novos casos após `add_product_to_consignment`:

- Remover produto restaura o estoque corretamente (verificar via `GET /produtos`)
- Remover produto faz com que ele não apareça mais em `GET /remessas/:id/produtos`
- Remover produto que não está na remessa retorna 404
- Consignatário (monitor) não consegue remover produto (403)
- Loja diferente não consegue remover produto da remessa alheia (403)

O teste de atomicidade (falha no incremento de estoque preserva o registro em packages) é considerado redundante — a atomicidade é garantida pelo PostgreSQL via `TransactionAware`, não sendo necessário um teste de integração específico.

## Out of Scope

- Remoção em lote de múltiplos produtos de uma vez
- Notificações ou auditoria sobre a remoção
- Soft-delete com status em vez de remoção física da linha
- Interface administrativa (fora do escopo do backend)
- Bloqueio de produtos duplicados em remessas (tratado no bean separado `pvapi-dnjn`)
- Limpeza de `package_reserved` em checkouts recebidos (lembrete para Kerlon)

## Further Notes

- A exclusão da remessa inteira (`DELETE /remessas/:consignment`) atualmente apenas arquiva (status `ARQUIVADO`) sem restaurar estoque. Isso é um bug/feature separado, fora do escopo deste PRD.
- O estoque do produto é decrementado no `AddProductAction.Exec` via `product_decrement_inventory`. A remoção faz o caminho inverso com `product_increment_inventory`.
- A FK `package_reserved → packages` tem `ON DELETE NO ACTION`, portanto o DELETE do pacote falharia se houvesse reservas. A verificação prévia de reservas evita esse erro em runtime.

\n\n## Summary of Changes\n\nImplemented DELETE /remessas/:consignment/produtos/:product endpoint to remove individual products from a consignment.\n\n### Changes made:\n\n1. **Migration 000033** — Created  PL/pgSQL function (symmetric to )\n2. **query/consignment.go** — Added queries: , , \n3. **query/product.go** — Added  query calling \n4. **util/util.go** — Added  and  sentinels,  handler for 409 Conflict responses\n5. **models/consignmentmodel/actions.go** — Created  struct with  method: validates non-archived, no active reservations, reads quantity, increments stock, deletes package line\n6. **controllers/consignment/consignment.go** — Added \n7. **gate/consignment.go** — Added  support in  endpoint block (calls  →  → )\n8. **routes.go** — Registered  route\n9. **api_consignment_test.go** — Added 5 tests: 404 for nonexistent product, 403 for consignee, 403 for different shop, 200 with success message, removed product not in listing
