---
# pvapi-gmru
title: Monitor — editar dados cadastrais e endereço
status: todo
type: feature
priority: normal
created_at: 2026-06-01T17:52:32Z
updated_at: 2026-06-01T17:53:24Z
---

## Problem Statement

Estabelecimentos (shops) podem criar monitores (consignees) com nome, email, telefone e endereço. Porém, não há endpoint para editar esses dados após a criação. Se um monitor mudar de endereço, telefone, ou precisar corrigir o nome/email, o estabelecimento precisa deletar e recriar o monitor, perdendo o vínculo com remessas e ocorrências existentes.

## Solution

Criar um endpoint `PUT /monitores/:id` que permita ao estabelecimento dono do monitor editar parcialmente os dados do monitor: nome, email, telefone e endereço (rua, número, bairro, cidade, estado, CEP, complemento). Todos os campos são opcionais — apenas os campos enviados são alterados. A operação é atômica: atualiza as tabelas `users`, `consignees` e `addresses` em uma única transação. O endpoint retorna o recurso completo do monitor atualizado.

## User Stories

1. Como **estabelecimento**, quero editar o nome de um monitor, para corrigir erros de digitação sem perder o vínculo com remessas ativas.
2. Como **estabelecimento**, quero editar o email de um monitor, para que ele possa fazer login com um email atualizado.
3. Como **estabelecimento**, quero editar o telefone de um monitor, para refletir mudanças de número de contato.
4. Como **estabelecimento**, quero editar o endereço completo de um monitor (rua, número, bairro, cidade, estado, CEP, complemento), para que os dados de entrega estejam sempre corretos.
5. Como **estabelecimento**, quero editar apenas um subconjunto dos campos (ex: só o telefone), sem precisar reenviar todos os dados, para simplificar a operação.
6. Como **estabelecimento**, quero editar apenas campos de endereço (ex: só o CEP), mantendo os demais dados do monitor inalterados.
7. Como **estabelecimento**, quero receber o monitor completo na resposta do update, para confirmar visualmente o novo estado dos dados.
8. Como **estabelecimento**, quero que a edição seja rejeitada se eu não for o dono do monitor, para evitar alterações indevidas por terceiros.
9. Como **estabelecimento**, quero receber um erro claro (404) se tentar editar um monitor que não existe, para tratar o erro no frontend adequadamente.
10. Como **estabelecimento**, quero que um PUT com corpo vazio (`{}`) seja aceito como no-op, retornando o monitor sem alterações, para que o frontend possa implementar salvamento condicional sem lógica adicional.
11. Como **monitor**, NÃO quero poder editar meus próprios dados via este endpoint — apenas o estabelecimento dono pode fazê-lo.
12. Como **admin**, NÃO quero poder editar dados de monitores de estabelecimentos alheios via este endpoint.

## Implementation Decisions

1. **Atualização parcial com campos opcionais**: Todos os campos do corpo da requisição são ponteiros (`*string`). Campos `nil` são ignorados; campos não-nil substituem o valor existente. Isso usa o padrão já adotado em `bankaccountmodel` e `pickupmodel`.

2. **Transação atômica entre 3 tabelas**: A atualização consulta `user_id` e `address_id` da tabela `consignees`, depois aplica `UPDATE` em `users` (nome, email), `consignees` (telefone) e `addresses` (todos os campos de endereço) dentro de `pgx.BeginFunc`. Se qualquer UPDATE falhar, toda a transação é revertida.

3. **Endereço editado in-place**: O endereço existente (linha na tabela `addresses`) é atualizado diretamente — não se cria um novo registro de endereço. Isso evita órfãos e mantém o `address_id` do monitor estável.

4. **Queries geradas dinamicamente com Squirrel**: As cláusulas `SET` são construídas com `squirrel.Update(table).SetMap(fields)` onde o mapa contém apenas os campos não-nil. Segue o mesmo padrão de `shopmodel.Update`. Tabelas sem alterações no payload são puladas (não geram UPDATE).

5. **Verificação de propriedade no handler**: Antes do update, o handler verifica `shopmodel.HasConsignee(ctx, pool, shopID, consigneeID)`. Se falso, retorna 404. Isso replica o padrão de `bankaccount.UpdateHandler` e `consignee.DeleteHandler`.

6. **Retorno do recurso completo**: Após o update bem-sucedido, o handler consulta `consignees_info_full_view` via `FindByID` e retorna o `ConsigneeResource` completo, mesmo padrão do `CreateHandler`.

7. **Novo tipo `AddressUpdateRequest`**: O pacote `address` ganha um struct `AddressUpdateRequest` com todos os campos como `*string`, paralelo ao existente `AddressRequest` (que tem `binding:"required"`). Ambos coexistem com propósitos distintos: criação vs atualização parcial.

8. **Estrutura da requisição**: O corpo JSON mantém o formato aninhado do endpoint de criação — `name`, `email`, `phone` no nível raiz e `address` como objeto aninhado. O campo `address` pode ser omitido ou enviado como `{}` quando nenhum campo de endereço muda.

9. **Autorização pelo gate existente**: O `ConsigneeGate.Modify` já está implementado e verifica propriedade via `Owns`. O handler adiciona verificação redundante de propriedade como defesa em profundidade.

10. **Rota**: `PUT /monitores/:consignee` é adicionado ao grupo `consignees` em `routes.go`. O parâmetro de URL usa o nome `:consignee` para consistência com DELETE.

## Testing Decisions

- **Testes de integração** no arquivo `api_monitor_test.go`, usando o harness `apiHarness` existente (que já provê `createMonitor`, `login`, `doJSON`, `assertStatus`).
- **O que torna um bom teste**: Verificar apenas comportamento externo (status HTTP, corpo da resposta), não detalhes de implementação (queries SQL, transações internas). Cada caso testa uma única condição de contorno.
- **Casos de teste**:
  - *Happy path*: criar monitor, fazer PUT parcial (apenas nome), verificar resposta com nome atualizado e demais campos inalterados
  - *Autorização/propriedade*: estabelecimento não-dono tenta editar → 403 ou 404
  - *Não encontrado*: PUT para ID inexistente → 404
  - *Corpo vazio*: PUT com `{}` → 200 com dados inalterados
- **Prior art**: Os testes seguem exatamente o padrão de `TestMonitor` e dos testes de banco (`api_saque_test.go` para updates).

## Out of Scope

- Edição de senha do monitor (escopo de redefinição de senha, endpoint separado)
- Edição de dados do monitor pelo próprio monitor (apenas estabelecimento dono)
- Troca de estabelecimento dono do monitor (transferência de ownership)
- Edição em lote de múltiplos monitores
- Histórico/auditoria de alterações
- Notificação ao monitor sobre alteração de dados

## Further Notes

- O endpoint de criação (`POST /monitores`) não é afetado por esta mudança.
- A view `consignees_info_full_view` já inclui todos os campos necessários para o retorno do recurso atualizado, sem necessidade de nova view ou migração.
- As queries de UPDATE são inteiramente geradas por Squirrel — nenhuma constante de query nova é necessária no pacote `query/consignee.go` além da `FindByID`.

Triage: ready-for-agent
