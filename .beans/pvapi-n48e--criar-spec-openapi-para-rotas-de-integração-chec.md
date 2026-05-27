---
# pvapi-n48e
title: Criar spec OpenAPI para rotas de integração (checkout, remessas, produtos)
status: in-progress
type: task
created_at: 2026-05-27T07:19:56Z
updated_at: 2026-05-27T07:19:56Z
---

Criar arquivo openapi spec com:
- Rotas do checkout (carrinho, pacotes, confirm)
- GET /remessas-publicas
- GET /estabelecimentos/{shop}/remessas
- GET /produtos com filtros shop e consignment
- GET /produtos/{product}
- Corrigir discrepância package_id no ConfirmCheckoutItem

Escrever em Dev/requests_implemented/api-integrations.yaml e buildar com Dev/build-docs.
