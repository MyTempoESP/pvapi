---
# pvapi-v985
title: Polir e endurecer o cliente de exemplo
status: todo
type: task
priority: normal
created_at: 2026-05-30T17:37:53Z
updated_at: 2026-05-30T17:37:58Z
parent: pvapi-outm
---

## Parent

pvapi-outm — Cliente de exemplo Streamlit

## What to build

Quatro pequenas correções de robustez no cliente de exemplo:

1. **Pin de dependências:** trocar `>=` por `==` no `requirements.txt` (versões exatas de streamlit e requests) para builds reprodutíveis.

2. **Query string via `params`:** na aba "Requisição HTTP", a query string atualmente é concatenada manualmente na URL (`path + "?" + query`). Trocar para parsing da string em dict e passar via `params=` do requests, garantindo encoding correto de caracteres especiais.

3. **Content-Type condicional:** a função `req()` define `Content-Type: application/json` em todas as requisições, inclusive GET e DELETE. Só aplicar o header em métodos com corpo (POST, PUT, PATCH).

4. **Usuário não-root no Dockerfile:** adicionar `RUN useradd --create-home appuser` e `USER appuser` após o `pip install`.

## Acceptance criteria

- [ ] `requirements.txt` usa `==` com versões exatas
- [ ] Query string com caracteres especiais (ex: `nome=São Paulo&valor=10%`) produz URL correta
- [ ] GET e DELETE não enviam header `Content-Type`
- [ ] POST, PUT, PATCH continuam enviando `Content-Type: application/json`
- [ ] Dockerfile cria e usa usuário não-root antes do CMD

## Blocked by

None — can start immediately.

Triage: ready-for-agent
