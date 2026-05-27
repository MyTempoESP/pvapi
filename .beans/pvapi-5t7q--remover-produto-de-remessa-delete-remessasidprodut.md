---
# pvapi-5t7q
title: Remover produto de remessa (DELETE /remessas/:id/produtos/:product)
status: todo
type: feature
priority: normal
created_at: 2026-05-27T15:49:56Z
updated_at: 2026-05-27T15:50:00Z
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
1. Remove a linha da tabela `packages` para aquele par (product_id, consignment_id)
2. Devolve a quantidade do produto ao estoque da loja (`products.stock`)
3. Respeita o mesmo controle de acesso do `PUT` (apenas a loja dona da remessa pode remover)

## User Stories

1. Como lojista, quero remover um produto específico de uma remessa, para corrigir erros de sortimento sem precisar recriar a remessa inteira.
2. Como lojista, quero que o estoque do produto seja restaurado automaticamente ao removê-lo da remessa, para não perder inventário.
3. Como lojista, quero receber um erro claro se tentar remover um produto que não está na remessa, para entender o que deu errado.
4. Como lojista, quero que apenas minha própria loja possa remover produtos das minhas remessas, para evitar manipulação por terceiros.
5. Como consignatário (monitor), NÃO quero poder remover produtos de remessas — apenas visualizá-los — mantendo a separação de responsabilidades atual.
6. Como lojista, quero que a remoção seja atômica (restaura estoque E remove o registro na mesma transação), para evitar inconsistências.

## Implementation Decisions

### Novo endpoint
- `DELETE /remessas/:consignment/produtos/:product` — recebe os dois IDs pela URL, sem body.

### Fluxo transacional
A operação executa dentro de uma transação:
1. Lê a quantidade atual do `packages` para o par (product_id, consignment_id)
2. Incrementa `products.stock` com essa quantidade (`product_increment_inventory`)
3. Remove a linha de `packages`

Isso é o inverso do `AddProductAction.Exec`, que decrementa estoque e insere em `packages`.

### Controle de acesso
Mesmo gate do `PUT /remessas/:consignment/produtos/:product`: apenas a loja dona (`ConsignmentGate.Owns` → `shopOwns`). O `ConsignmentGate.Dispatch` já tem um case para `/produtos/:product`, basta adicionar `r.Method == "DELETE"` retornando `c.Modify(u, r)`.

### Restauração de estoque
Criar função `productmodel.ReturnToStock(ctx, tx, productID, qty)` que chama `product_increment_inventory` (nova função SQL ou reuso com valor negativo). Preferível criar uma função explícita `product_increment_inventory` para clareza.

### Erro se produto não está na remessa
Se a query de leitura da quantidade retornar `pgx.ErrNoRows`, retornar 404 com mensagem apropriada.

## Testing Decisions

### O que testar
- Remover produto restaura o estoque corretamente (verificar via `GET /produtos`)
- Remover produto faz com que ele não apareça mais em `GET /remessas/:id/produtos`
- Remover produto que não está na remessa retorna 404
- Consignatário (monitor) não consegue remover produto (403)
- Loja diferente não consegue remover produto da remessa alheia (403)
- Operação é atômica: se a restauração de estoque falhar, o registro em packages permanece

### Estilo de teste
Seguir o padrão de `api_consignment_test.go`: harness-based, subtests com `t.Run`, verificação de status code e body JSON. Adicionar os novos casos após `add_product_to_consignment`.

## Out of Scope

- Remoção em lote de múltiplos produtos de uma vez
- Notificações ou auditoria sobre a remoção
- Soft-delete com status em vez de remoção física da linha
- Interface administrativa (fora do escopo do backend)

## Further Notes

- A exclusão da remessa inteira (`DELETE /remessas/:consignment`) atualmente apenas arquiva (status ARQUIVADO) sem restaurar estoque. Isso é um bug/feature separado, fora do escopo deste PRD.
- O estoque do produto é decrementado no `AddProductAction.Exec` via `product_decrement_inventory`. A remoção deve fazer o caminho inverso com uma função simétrica.

Triage: ready-for-agent
