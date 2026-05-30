---
# pvapi-h0sa
title: Corrigir quantidade na confirmação do checkout
status: scrapped
type: bug
priority: normal
created_at: 2026-05-30T17:36:54Z
updated_at: 2026-05-30T17:37:34Z
---

## Parent

pvapi-outm — Cliente de exemplo Streamlit

## What to build

No fluxo de Checkout, a Seção 4 (Confirmar compra) envia a  dos  lendo da variável  da Seção 2 (campo de quantidade do carrinho) usando o padrão frágil . Isso mistura dois valores que não têm relação: quantas unidades foram adicionadas ao carrinho vs quantas unidades confirmar no pedido.

**Correção:** adicionar um input dedicado "Quantidade" (number_input, default 1, min 1) na Seção 4. O payload de confirmação usa esse valor. Remover completamente a referência ao  e o  anti-pattern.

## Acceptance criteria

- [ ] Seção 4 do Checkout exibe campo "Quantidade" próprio (default 1, min 1)
- [ ] Payload de confirmação usa o valor desse campo, não 
- [ ] Expressão  removida do código
- [ ] Fluxo completo funciona: carregar produtos → adicionar ao carrinho → carregar pacotes → selecionar pacote → definir quantidade → confirmar

## Blocked by

None — can start immediately.
