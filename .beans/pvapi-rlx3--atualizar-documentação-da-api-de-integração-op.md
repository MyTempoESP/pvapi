---
# pvapi-rlx3
title: Atualizar documentação da API de integração (OpenAPI + llms.txt) e remover session_id do checkout
status: todo
type: feature
priority: normal
created_at: 2026-06-05T01:29:55Z
updated_at: 2026-06-05T01:29:59Z
---

## Problem Statement

A documentação pública da API de integração do PvAuto — o schema OpenAPI em `api-integrations.yaml` e o texto legível por LLMs em `llms.txt` — está seriamente desatualizada. Ela descreve um fluxo baseado em carrinho (adicionar ao carrinho → listar pacotes → confirmar checkout) que não existe mais no código. As rotas `/carrinho/*` foram removidas, os query params renomeados, e novos campos (como `pickups` inline nos produtos) não aparecem.

Além disso, o endpoint `POST /checkouts/confirm` ainda exige um `session_id` obrigatório que é vestigial — ele só aparece na coluna `description` do banco, sem função real. O parâmetro de idempotência real é `external_reference`.

Integradores que consultam a documentação atual recebem informações incorretas sobre como consumir a API.

## Solution

1. **Remover `session_id` do checkout (código):** Tornar o parâmetro obsoleto no controller e no modelo — remover a validação de obrigatoriedade e não passá-lo para a query de inserção.
2. **Reescrever `api-integrations.yaml`:** Remover tag Carrinho e todas as rotas `/carrinho/*`. Atualizar `/produtos` com os query params corretos (`loja`, `remessa`) e documentar o campo `pickups` em cada produto. Atualizar o fluxo descrito no topo para refletir o modelo cartless (browse produtos com pickups → confirm checkout). Atualizar schemas de produto e checkout.
3. **Reescrever `llms.txt`:** Mesmas mudanças, em formato texto legível por LLMs.

A documentação final deve descrever fielmente a API pública real:
- `GET /produtos` — listagem com filtros `?loja=` e `?remessa=`, com `pickups` inline
- `GET /produtos/:product` — detalhe com `pickups` inline
- `GET /remessas-publicas` — lista de remessas ativas
- `GET /estabelecimentos/:shop/remessas` — remessas por loja
- `GET /estabelecimentos` — lista de lojas ativas
- `POST /checkouts/confirm` — confirmação de checkout (sem `session_id` obrigatório), com criação automática de cliente via `client_email`/`client_name`

## User Stories

1. Como integrador, quero consultar a documentação OpenAPI e ver exatamente os endpoints que existem, para não perder tempo tentando usar rotas removidas.
2. Como integrador, quero entender o fluxo de compra real (browse produtos → confirm checkout), para implementar a integração corretamente.
3. Como integrador, quero saber que cada produto já inclui os pacotes/pickups disponíveis (preço, estoque, endereço do ponto de retirada), para selecionar o `package_id` diretamente sem etapa intermediária de carrinho.
4. Como integrador, quero saber os query params corretos para filtrar produtos (`loja` e `remessa`), para usar os nomes que a API realmente aceita.
5. Como integrador, quero que a documentação reflita os campos reais da resposta de produto (incluindo `pickups` e sem o campo obsoleto `consignments`), para fazer parsing correto no meu código.
6. Como integrador, quero saber que o checkout aceita `client_email` e `client_name` para criação automática de conta, para oferecer essa funcionalidade aos meus usuários.
7. Como integrador usando LLMs para gerar código de integração, quero que o `llms.txt` esteja atualizado, para que o código gerado seja funcional.
8. Como desenvolvedor da plataforma, quero remover o `session_id` obrigatório do checkout, para eliminar um parâmetro vestigial que confunde integradores.
9. Como integrador, quero saber que o checkout é idempotente via `external_reference`, para fazer retries seguros sem duplicar transações.
10. Como integrador, quero ver schemas de resposta precisos para erros HTTP (400, 404, 409, 422), para tratar cada caso adequadamente.

## Implementation Decisions

### Remoção de session_id

- Remover validação de obrigatoriedade (`sessionID == ""`) no controller `checkout.ConfirmHandler`
- Remover `sessionID` da assinatura de `checkoutmodel.Confirm`
- Remover uso de `sessionID` na montagem da `description` em `writeCheckout` — substituir por descrição genérica ou usar `external_reference`
- O parâmetro de idempotência continua sendo `external_reference`

### Documentação OpenAPI (api-integrations.yaml)

- Remover tag `Carrinho` e todos os paths `/carrinho/*`
- Remover schemas: `AdicionarCarrinhoRequest`, `ItemCarrinho`, `RespostaCarrinho`, `OpcaoPacote`, `PacotesPorProduto`, `RespostaPacotes`, `PontoRetirada`
- Atualizar `/produtos`: query params `loja` (integer[], repeatable) e `remessa` (integer[], repeatable)
- Adicionar schema `ProductPickupResource` com `package_id`, `price`, `discount`, `quantity`, `pickup_name`, `pickup_date`, `address` (Endereco aninhado)
- Adicionar `pickups` como array de `ProductPickupResource` no `RecursoProduto`
- Remover campo `consignments` do `RecursoProduto`
- Atualizar `ConfirmarCheckoutRequest`: remover `session_id` dos query params. Manter `value_paid`, `external_reference`, `items`, `client_email`, `client_name`
- Atualizar descrição do fluxo principal no `info.description` para refletir o modelo cartless
- Manter endpoints: `/produtos`, `/produtos/:product`, `/remessas-publicas`, `/estabelecimentos/:shop/remessas`, `/estabelecimentos`, `/checkouts/confirm`

### Documentação llms.txt

- Mesmas mudanças estruturais que o YAML, em formato markdown legível por LLMs
- Manter consistência de nomenclatura e exemplos com o YAML
- Incluir exemplo de request/response para cada endpoint

## Testing Decisions

- Testes devem verificar apenas comportamento externo (HTTP status codes, estrutura JSON das respostas)
- Os testes existentes em `api_checkout_client_test.go` e `api_product_test.go` já cobrem o comportamento real
- Após a remoção de `session_id`, executar a suíte de testes completa via skill `test-code` para garantir que nada quebrou
- Se testes existentes dependem do `session_id` como obrigatório, atualizá-los para refletir o novo comportamento

## Out of Scope

- `api-integrations.md` — não será alterado
- `POST /payments/session/validate` — não é um endpoint real de integração
- `GET /estabelecimentos/:shop/produtos` — não inclui pickups, portanto não entra na documentação de integração
- Geração automática de docs a partir do YAML (tooling)
- Alterações nos endpoints autenticados (`/estabelecimentos` com auth, etc.)
- Refatoração de outros arquivos YAML em `requests_implemented/`

## Further Notes

- O `session_id` era usado apenas para montar a string `description` na tabela `checkouts`. Removê-lo não afeta idempotência nem lógica de negócio.
- A criação automática de cliente (`client_email`/`client_name`) é best-effort e acontece após o envio da resposta HTTP — não afeta o status do checkout.
- O campo `pickups` em produtos vem de `BatchFetchPickups`, que faz join entre packages, consignments, e pickup_locations.
- O `index.html` gerado pelo Redocly (`build-docs`) será reconstruído automaticamente a partir do YAML atualizado.

Triage: ready-for-agent
