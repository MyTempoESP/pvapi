---
# pvapi-gezx
title: Corrigir quantidade na confirmação do checkout
status: todo
type: bug
priority: normal
created_at: 2026-05-30T17:37:53Z
updated_at: 2026-05-30T17:37:58Z
parent: pvapi-outm
---

## Parent

pvapi-outm — Cliente de exemplo Streamlit

## What to build

No fluxo de Checkout, a Seção 4 (Confirmar compra) envia a `quantity` dos `items` lendo da variável `add_qty` da Seção 2 (campo de quantidade do carrinho) usando o padrão frágil `add_qty if "add_qty" in dir() else 1`. Isso mistura dois valores que não têm relação: quantas unidades foram adicionadas ao carrinho vs quantas unidades confirmar no pedido.

**Correção:** adicionar um input dedicado "Quantidade" (number_input, default 1, min 1) na Seção 4. O payload de confirmação usa esse valor. Remover completamente a referência ao `add_qty` e o `dir()` anti-pattern.

## Acceptance criteria

- [ ] Seção 4 do Checkout exibe campo "Quantidade" próprio (default 1, min 1)
- [ ] Payload de confirmação usa o valor desse campo, não `add_qty`
- [ ] Expressão `add_qty in dir()` removida do código
- [ ] Fluxo completo funciona: carregar produtos → adicionar ao carrinho → carregar pacotes → selecionar pacote → definir quantidade → confirmar

## Blocked by

None — can start immediately.

Triage: ready-for-agent
