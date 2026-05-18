# Histórias Técnicas (Backlog Refinado)

Baseado no relatório de esforço (`report.md`) e no código-fonte real.

---

## Sprint 1 – Base, correções, pickups

### 1. [Bug] Logar como monitor  
---  
**Tempo médio:** 4 h (c/ folga)  
**Descrição:**  
Corrigir bug que impede o login de usuários com perfil "monitor" (consignee). Ajustar o modelo `authmodel` para incluir as claims corretas no JWT e modificar a rota de autorização em `controllers/policy/policy.go` para permitir o acesso.

**Arquivos tocados:**  
- `models/authmodel/authmodel.go`  
- `controllers/policy/policy.go`  

**Detalhes:**  
- Reproduzir o bug e identificar falha na geração/validação do token para o role `consignee`.  
- Adicionar o campo `role` nas claims do JWT dentro da função `authenticator`.  
- Atualizar o bouncer de policy para reconhecer o prefixo `/monitores` e liberar o login.  

---

### 2. [Bug] Criar ocorrência (formato front)  
---  
**Tempo médio:** 3 h (c/ folga)  
**Descrição:**  
O endpoint de criação de ocorrências (`/ocorrencias`) não está parseando corretamente o JSON enviado pelo front‑end. Ajustar o handler `controllers/report/report.go` para aceitar o formato real (ex.: `occurrenceScope`, `orderId` etc.) e mapear para a estrutura `CreateReportRequest`.

**Arquivos tocados:**  
- `controllers/report/report.go` (handler `CreateHandler` – ajustar parse)  
- `models/reportmodel/requests.go` (estrutura `CreateReportRequest` – adicionar tags `json` corretas)  

**Detalhes:**  
- O front‑end envia campos como `occurrenceScope` e `orderId`, enquanto a API espera `scope` e `consignment_id`.  
- Adicionar tags `json` correspondentes nos campos da struct ou criar um alias de parse.  
- Testar a rota com o payload real do front.  

---

### 3. [$easy] Listar clientes  
---  
**Tempo médio:** 4 h (c/ folga)  
**Descrição:**  
Criar endpoint `GET /clientes` para listar os clientes de uma loja. Já existe a query `ListShopClientsQuery` em `query/client.go` e o método `List` em `models/clientmodel/model.go`. Basta expor um handler em `controllers/client/client.go`.

**Arquivos tocados:**  
- `query/client.go` (já pronto)  
- `models/clientmodel/model.go` (método `List` já implementado)  
- **`controllers/client/client.go`** (criar handler `ListHandler`)  

**Detalhes:**  
- Reutilizar `util.WithCollect` para retornar os clientes associados ao shop logado.  
- Registrar a rota `$GET /clientes` no arquivo `routes.go` dentro do grupo `withAuth`.  

---

### 4. [$easy] Unificar cadastro com `FindOrCreate`  
---  
**Tempo médio:** 5 h (c/ folga)  
**Descrição:**  
Substituir as chamadas diretas a `user.Save` nos métodos de criação de loja, monitor e cliente por uma função `FindOrCreate` (já existente em `models/usermodel/model.go`). Isso evita duplicatas de usuário quando o mesmo e‑mail é usado.

**Arquivos tocados:**  
- `models/shopmodel/model.go` (método `CreateTx`)  
- `models/consigneemodel/model.go` (método `Save`)  
- `models/clientmodel/model.go` (método `CreateForCart`)  

**Detalhes:**  
- A função `FindOrCreate` faz um `INSERT ... ON CONFLICT` e retorna o `id` do usuário existente ou recém‑criado.  
- Alterar `user.Save` para `user.FindOrCreate` onde aplicável, ajustando a assinatura e passando os parâmetros corretos (name, email, passwordHash).  
- Remover a checagem prévia de existência (se houver).  

---

### 5. [$easy] Alterar queries de inserção para upsert  
---  
**Tempo médio:** 4 h (c/ folga)  
**Descrição:**  
Adicionar cláusulas `ON CONFLICT` nas queries de inserção de lojas, monitores e clientes para evitar erros de duplicidade. As queries atuais lançam erro se já existir um registro com a mesma chave única.

**Arquivos tocados:**  
- `query/shop.go`  
- `query/consignee.go`  
- `query/client.go`  

**Detalhes:**  
- Em `ShopCreateQuery`, adicionar `ON CONFLICT (slug) DO NOTHING`.  
- Em `CreateConsigneeQuery`, adicionar `ON CONFLICT (user_id, shop_id) DO NOTHING`.  
- Em `CreateClientQuery` (já possui `ON CONFLICT`), verificar se está correto.  
- Testar inserções repetidas.  

---

### 6. [$easy] Criar tabela `package_sale`  
---  
**Tempo médio:** 4 h (c/ folga)  
**Descrição:**  
Criar uma nova tabela `package_sale` no banco de dados para registrar a venda de pacotes (itens de remessa). Isso envolve escrever um script SQL de migração e definir o struct correspondente em Go.

**Arquivos tocados:**  
- `models/consignmentmodel/data.go` (novo struct `PackageSale`)  
- **Migração SQL** (criar tabela `package_sale`)  

**Detalhes:**  
- Tabela sugerida: `package_sale (id, package_id, sale_id, price_sold, quantity_sold, created_at)`.  
- Criar a migração no diretório `internal/database/migrations/`.  
- Definir struct com tags `db` para mapeamento.  

---

### 7. [$easy] Filtragem por data (pickup)  
---  
**Tempo médio:** 3 h (c/ folga)  
**Descrição:**  
Modificar a query de listagem de pontos de retirada (`pickup_locations`) para retornar apenas aqueles com data futura (agendados para depois de agora). Isso evita exibir pickups vencidos.

**Arquivos tocados:**  
- `query/pickup.go` (adicionar `WHERE pickup_at >= NOW()`)  

**Detalhes:**  
- Na constante `ListPickupLocationsQuery`, adicionar cláusula `AND p.pickup_at >= NOW()`.  
- Opcional: ordenar por data crescente.  

---

### 8. [$easy] Listagem simples de pickup (dropdown)  
---  
**Tempo médio:** 3 h (c/ folga)  
**Descrição:**  
Criar um endpoint ou reutilizar o existente `GET /pontos` para retornar apenas os campos necessários (id, name, datetime) para preenchimento de dropdown no front‑end. Já existe a função `List` em `models/pickupmodel/model.go`.

**Arquivos tocados:**  
- `models/pickupmodel/model.go` (método `List` já pronto)  
- `controllers/pickups/pickups.go` (criar handler simplificado ou ajustar existente)  

**Detalhes:**  
- Se o endpoint atual já retorna todos os dados, pode ser suficiente.  
- Caso contrário, criar um novo handler que retorne apenas uma versão resumida.  

---

### 9. [$average] Criar struct `POSRequest{}`  
---  
**Tempo médio:** 7 h (c/ folga)  
**Descrição:**  
Criar uma nova estrutura `POSRequest` para representar uma requisição de ponto de venda (Point of Sale – pickup no ato da compra). Adicionar o campo `PickupLocationID` no pacote de criação de remessa.

**Arquivos tocados:**  
- `models/consignmentmodel/requests.go` (adicionar campo `PickupLocationID` e struct)  

**Detalhes:**  
- `POSRequest` deve conter `PickupLocationID int` e possivelmente outros campos como `monitorId`.  
- Adaptar o método `Prepare` para mapear os campos corretamente.  

---

### 10. [$average] Modificar controller para aceitar POS  
---  
**Tempo médio:** 7 h (c/ folga)  
**Descrição:**  
Atualizar os controllers de carrinho e remessa para aceitar o novo campo `POSRequest` durante o checkout e criação de remessa. Ajustar os handlers `controllers/cart/cart.go` e `controllers/consignment/consignment.go` para parsear e utilizar o ID do ponto de retirada.

**Arquivos tocados:**  
- `controllers/cart/cart.go`  
- `controllers/consignment/consignment.go`  

**Detalhes:**  
- No fluxo de finalização do carrinho (`CartPIXHandler`), receber `PickupLocationID` no JSON e passá-lo adiante.  
- Na criação de remessa, modificar `CreateConsignmentRequest` para aceitar `destinationId` (já existe) e garantir que o campo seja utilizado corretamente.  

---

## Sprint 2 – Checkout, frete, modal

### 11. [$hard] Modificar funções de checkout para usar referência a pickup  
---  
**Tempo médio:** 21 h (c/ folga)  
**Descrição:**  
Reestruturar o fluxo de checkout para que a remessa gerada após o pagamento referencie o ponto de retirada (pickup) escolhido pelo cliente. Isso exige alterações no `cartmodel`, nas queries de finalização, no controller de carrinho e no modelo de remessa.

**Arquivos tocados:**  
- `models/cartmodel/model.go` (método `HoldStock` e `Exec`)  
- `query/cart.go` (`CartHoldStock`)  
- `controllers/cart/cart.go` (fluxo de finalização)  
- `models/consignmentmodel/model.go` (`AddProductAction.Exec`)  

**Detalhes:**  
- Durante `HoldStock`, capturar o `pickup_location_id` selecionado e armazená-lo na sessão ou no cart.  
- Ao criar a shipment (ou consignment) pós‑pagamento, ler esse ID e associá-lo à remessa.  
- Garantir que as transações sejam atômicas (rollback em caso de erro).  

---

### 12. [$medium] Frete simplificado  
---  
**Tempo médio:** 10 h (c/ folga)  
**Descrição:**  
Implementar cálculo de frete simplificado para MVP (ex.: valor fixo de R$15). Adicionar uma função em `util/util.go` e incluir o campo `freightCost` na requisição de pagamento.

**Arquivos tocados:**  
- `util/util.go` (função de cálculo fixo/tabela)  
- `models/paymentmodel/requests.go` (adicionar campo `freightCost`)  

**Detalhes:**  
- Criar função `CalculateFreight(items []CartProduct) decimal.Decimal` que retorna valor fixo.  
- Incluir o custo no total do pagamento e armazenar no banco.  

---

### 13. [$average] Modal de endereço da remessa  
---  
**Tempo médio:** 7 h (c/ folga)  
**Descrição:**  
Criar um endpoint `GET /remessas/:id/endereco` que retorna os dados de endereço associados à remessa (se for shipment, o endereço do cliente; se for consignment, o endereço do ponto de retirada). O struct já possui endereço, falta serializá‑lo.

**Arquivos tocados:**  
- `controllers/consignment/` (criar handler)  
- `models/consignmentmodel/data.go` (endereço já disponível no struct)  

**Detalhes:**  
- Verificar `record_type` para decidir qual endereço retornar.  
- Serializar no formato esperado pelo front‑end.  

---

## Sprint 3 – Estados, pagamentos, resumo + Financeiro

### 14. [$medium] Estados de entrega (PENDENTE/RETIRADO/DEVOLVIDO)  
---  
**Tempo médio:** 13 h (c/ folga)  
**Descrição:**  
Implementar máquina de estados para as remessas, permitindo alterar o status entre `PENDENTE SEPARAÇÃO`, `RETIRADO` e `DEVOLVIDO`. Criar enum, métodos de atualização, queries e verificar permissões no gate.

**Arquivos tocados:**  
- `models/consignmentmodel/data.go` (criar enum ou constantes)  
- `models/consignmentmodel/model.go` (métodos `SetStatus`)  
- `query/consignment.go` (UPDATE de status)  
- `gate/consignment.go` (verificar permissões para alterar status)  

**Detalhes:**  
- Definir constantes para os status.  
- Criar função `SetStatus(ctx, tx, id, newStatus) error`.  
- Adicionar validação de transições válidas.  

---

### 15. [$average] Listar pagamentos do cliente  
---  
**Tempo médio:** 7 h (c/ folga)  
**Descrição:**  
Criar endpoint `GET /clientes/:id/pagamentos` que retorna os pagamentos realizados por um determinado cliente. Criar nova query `ListByClient`, struct `PaymentResource` e handler.

**Arquivos tocados:**  
- `query/payment.go` (nova query `ListByClient`)  
- `models/paymentmodel/` (criar `PaymentResource`)  
- `controllers/client/` (handler)  

**Detalhes:**  
- Query deve filtrar por `client_id` (via `users` ou `clients`).  
- Resource deve conter `payment_id`, `amount`, `status`, `created_at`.  

---

### 16. [$average] Resumo de compras  
---  
**Tempo médio:** 7 h (c/ folga)  
**Descrição:**  
Criar endpoint que retorna o histórico de compras (carrinhos finalizados) de um cliente. Aproveitar as queries existentes em `query/cart.go` e criar método no `models/cartmodel`.

**Arquivos tocados:**  
- `query/cart.go` (histórico por cliente)  
- `models/cartmodel/model.go` (método `FindByClient`)  
- `controllers/cart/` (endpoint)  

**Detalhes:**  
- Filtrar carrinhos onde `session_id` começa com "FINISHED_".  
- Agrupar por `cart_id` e incluir lista de itens.  

---

### 17. [$hard] Geração e verificação de código QR  
---  
**Tempo médio:** 21 h (c/ folga)  
**Descrição:**  
Gerar código QR para cada remessa, permitindo verificação de status via leitura. Inclui criação de endpoint de geração e verificação, armazenamento na tabela de remessas.

**Arquivos tocados:**  
- `models/consignmentmodel/` (novo campo `qr_code`)  
- `controllers/qrcode/` (novo controller)  
- `query/consignment.go` (UPDATE com QR)  

---

### 18. [$medium] Criar modelo de saque (tabela, queries, estrutura Go)  
---  
**Tempo médio:** 10 h (c/ folga)  
**Descrição:**  
Criar a tabela `withdrawals` (saques) no banco de dados, com queries de inserção e listagem, structs em Go e modelo correspondente.

**Arquivos tocados:**  
- `query/withdrawal.go` (novo arquivo)  
- `models/withdrawalmodel/model.go` (novo)  
- Migration SQL correspondente  

---

### 19. [$medium] Endpoint solicitar saque (criação, validação de saldo)  
---  
**Tempo médio:** 10 h (c/ folga)  
**Descrição:**  
Criar endpoint `POST /saques` para lojistas solicitarem saque do saldo disponível. Incluir validação de saldo e atualização do saldo após criação.

**Arquivos tocados:**  
- `controllers/withdrawal/withdrawal.go` (novo handler)  
- `models/withdrawalmodel/model.go` (método `Create`)  
- `models/shopmodel/model.go` (consulta de saldo)  

---

### 20. [$average] Listar saques do usuário (histórico)  
---  
**Tempo médio:** 7 h (c/ folga)  
**Descrição:**  
Criar endpoint `GET /saques` que retorna histórico de solicitações de saque (filtro por loja) com status e datas.

**Arquivos tocados:**  
- `controllers/withdrawal/withdrawal.go` (handler `List`)  
- `models/withdrawalmodel/model.go` (método `ListByShop`)  

---

### 21. [$average] Painel financeiro: totais (receita líquida, comissões, saldo disponível)  
---  
**Tempo médio:** 8 h (c/ folga)  
**Descrição:**  
Criar endpoint `GET /painel/financeiro` que retorna resumo de totais financeiros do lojista: receita líquida (vendas realizadas), comissões pagas, saldo disponível para saque.

**Arquivos tocados:**  
- `controllers/finance/finance.go` (novo controller)  
- `models/paymentmodel/model.go` (agregações de vendas)  
- `models/withdrawalmodel/model.go` (total sacado)  

---

### 22. [$easy] Endpoint de aprovação/rejeição de saque (admin)  
---  
**Tempo médio:** 4 h (c/ folga)  
**Descrição:**  
Criar endpoint `PUT /saques/:id` para administrador aprovar ou rejeitar um saque pendente. Alterar status e, se aprovado, marcar saldo como debitado.

**Arquivos tocados:**  
- `controllers/withdrawal/withdrawal.go` (handler `UpdateStatus`)  
- `models/withdrawalmodel/model.go` (método `UpdateStatus`)  
- `gate/admin.go` (permissões de admin)  

---

**Nota:** Todos os tempos incluem 30% de folga (conforme report.md).  
**Dependências críticas** e alocação por sprint permanecem conforme planejamento original.
