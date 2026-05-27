---
# pvapi-dnjn
title: Bloquear produtos duplicados em remessas
status: done
type: feature
priority: normal
created_at: 2026-05-27T15:55:32Z
updated_at: 2026-05-27T15:55:32Z
blocking:
    - pvapi-5t7q
---

## Problem

Atualmente é possível adicionar o mesmo produto várias vezes na mesma remessa com parâmetros diferentes (cor, preço etc). Isso gera múltiplas linhas em packages com o mesmo (product_id, consignment_id).

## Solution

Validar no POST /remessas/:consignment/produtos se o product_id já existe naquela remessa. Se existir, retornar 409 Conflict.

## Dependencies

Bloqueia pvapi-5t7q (remover produto de remessa), que usa a estratégia de deletar todas as linhas do par (product_id, consignment_id).
