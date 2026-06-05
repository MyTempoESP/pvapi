---
# pvapi-ft3u
title: Substituir slug por nome formatado no orderName de relatórios com escopo consignment
status: completed
type: feature
priority: normal
created_at: 2026-06-05T04:46:52Z
updated_at: 2026-06-05T04:50:27Z
---

## Problem Statement

O campo `orderName` nos relatórios de ocorrências retorna o slug da remessa (`rem-generica`) para relatórios com escopo `consignment`. Isso não é informativo para o usuário — o slug é um valor interno genérico. O usuário precisa identificar rapidamente qual remessa o relatório se refere, com o nome do monitor e os pontos de retirada.

## Solution

Para relatórios com escopo `consignment`, o `orderName` passa a usar o formato:

> **{monitor} para {pickup1}, {pickup2}, {pickup3}, +N outros**

Onde:
- `monitor` é o nome do consignee associado ao relatório
- `pickup1`, `pickup2`, `pickup3` são os 3 primeiros nomes de pontos de retirada da remessa, ordenados alfabeticamente
- `+N outros` aparece apenas quando há mais de 3 pickups, onde N = total - 3

### Comportamento por número de pickups

| Pickups | Exemplo de orderName |
|---------|---------------------|
| 0 | `N/A` |
| 1 | `João Silva para Loja Centro` |
| 2 | `João Silva para Loja Centro, Loja Norte` |
| 3 | `João Silva para Loja Centro, Loja Norte, Loja Sul` |
| 5 | `João Silva para Loja Centro, Loja Norte, Loja Sul, +2 outros` |

Relatórios com escopo `product` continuam usando o nome do produto, sem alterações.

## User Stories

1. Como lojista, ao listar ocorrências, quero ver no campo orderName o monitor e os pontos de retirada da remessa, para identificar rapidamente qual remessa está com problema.
2. Como lojista, ao visualizar uma ocorrência de remessa, quero ver o monitor e pontos de retirada no orderName, para saber o contexto sem precisar abrir a remessa.
3. Como monitor, ao listar minhas ocorrências, quero identificar qual remessa cada ocorrência pertence pelo nome do monitor e pontos de retirada.
4. Como monitor, ao visualizar uma ocorrência de remessa, quero ver os pontos de retirada no orderName, para confirmar de qual ponto veio o produto reportado.
5. Como usuário da API, quero que o orderName seja informativo mesmo quando a remessa não tem pontos de retirada cadastrados, exibindo "N/A" em vez de um slug genérico.
6. Como usuário da API, quero que o orderName mostre até 3 pontos de retirada com um contador de restantes, para não sobrecarregar a interface com muitos nomes.

## Implementation Decisions

### Decisão 1: Cálculo em Go, não em SQL

O `orderName` para escopo consignment será computado em Go após as queries SQL, em vez de usar funções SQL inline. Motivo: SQL com string_agg e lógica condicional fica complexo e difícil de testar isoladamente. A lógica em Go é mais legível e testável.

### Decisão 2: Módulo de formatação puro

Uma função pura `BuildConsignmentOrderName(monitor string, pickups []Pickup) string` encapsula toda a lógica de formatação. Sem dependências de banco, trivial de testar com diferentes cenários de entrada.

### Decisão 3: Batch-fetch de pickups via consignmentmodel existente

As funções `ListByShop`, `ListByConsignee` e `Find` coletam os IDs das remessas com escopo consignment e chamam `consignmentmodel.BatchFetchConsignmentPickups` para buscar os pickups em uma única query. Isso evita N+1 queries.

### Decisão 4: SQL permanece inalterado

O `COALESCE(p.name, co.slug)` no SQL continua existindo como fallback. O valor é sobrescrito em Go para relatórios de escopo consignment.

### Decisão 5: Separador vírgula-espaço, sem "e" no último

Formato: `"PickupA, PickupB, PickupC"`. Sem conjunção no último elemento — mais simples e consistente com o sufixo `+N outros`.

### Decisão 6: Comportamento com 0 pickups

Retorna `"N/A"`. Sem o prefixo "monitor para", pois não há pontos de retirada para listar.

### Decisão 7: +N aparece apenas quando N > 0

Para 1-3 pickups, lista todos sem sufixo. Para 4+, mostra 3 nomes + `"+N outros"`.

## Testing Decisions

### O que testar

Apenas comportamento externo — entrada e saída da função de formatação, e o valor final do `orderName` nas respostas da API.

### Testes unitários (função pura)

Arquivo novo: testes para `BuildConsignmentOrderName` cobrindo:
- 0 pickups → "N/A"
- 1 pickup → "Monitor para Pickup1"
- 2 pickups → "Monitor para Pickup1, Pickup2"
- 3 pickups → "Monitor para Pickup1, Pickup2, Pickup3"
- 5 pickups → "Monitor para Pickup1, Pickup2, Pickup3, +2 outros"

### Testes de integração

No `api_report_test.go` existente:
- Criar remessa com múltiplos pickups e verificar `orderName` no response
- Criar remessa sem pickups e verificar `orderName` = "N/A"
- Verificar que `orderName` no list (GET /ocorrencias) também usa o novo formato
- Verificar que relatórios com escopo product continuam usando nome do produto (não regredir)

## Out of Scope

- Alterar o `orderName` para escopo `product` (continua usando nome do produto)
- Alterar o campo `slug` da remessa em si
- Expor os nomes de pickup como campo separado na resposta do relatório
- Internacionalização do formato

## Further Notes

- O slug `rem-generica` é hardcoded no `CreateConsignmentQuery`. Isso não é alterado neste PRD.
- `BatchFetchConsignmentPickups` já retorna pickups ordenados por `pl.name` (via SQL `ORDER BY pl.name`), garantindo consistência na ordem de exibição.



Triage: ready-for-agent



## Summary of Changes

Implementada a substituição do slug pelo nome formatado no orderName de relatórios com escopo consignment.

### Arquivos criados
- **models/reportmodel/ordername.go**: Função pura  que formata o orderName conforme a spec (ex: "João Silva para Loja Centro, Loja Norte, Loja Sul, +2 outros").
- **models/reportmodel/ordername_test.go**: Testes unitários para todos os cenários (0, 1, 2, 3, 5 pickups).

### Arquivos modificados
- **models/reportmodel/model.go**: Adicionados helpers  e  que fazem batch-fetch dos pickups via  e sobrescrevem o  para reports com scope consignment. Modificadas ,  e  para usar esses helpers.
- **api_report_test.go**: Adicionados 3 testes de integração: (1) consignment com 3 pickups verifica formato no view e list, (2) consignment sem pickups verifica "N/A", (3) product scope continua usando nome do produto (regressão).
- **api_fixtures_test.go**: Adicionado helper  para criar remessas com múltiplos pickups.
- **api_assert_test.go**: Adicionado helper  para validação do formato do orderName.

### Decisões de implementação seguidas
- Cálculo em Go puro, não em SQL
-  para evitar N+1
- SQL inalterado (COALESCE continua como fallback)
- Formato com vírgula-espaço, sem "e" no último elemento
- 0 pickups → "N/A"
