# Contas Bancárias e Saques (Bank Accounts & Withdrawals)

A loja (role `shop`) gerencia suas **contas bancárias** e solicita **saques** vinculados a essas contas.  
Todas as rotas exigem autenticação JWT com role `shop`.

---

## Fluxo geral

```
1. CRUD contas bancárias     ← criar/listar/editar/excluir contas de destino
2. POST /estabelecimentos/saques  ← criar saque referenciando bank_account
3. GET  /estabelecimentos/saques  ← listar saques
4. GET  /estabelecimentos/saques/:id ← ver detalhes
5. PUT  /estabelecimentos/saques/:id ← aprovar/rejeitar/cancelar
6. DELETE /estabelecimentos/saques/:id ← remover
```

---

## Contas Bancárias

### 1. Criar conta bancária

```
POST /estabelecimentos/contasbancarias
Authorization: Bearer <token>
```

**Body:**

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `name` | string | sim | Apelido da conta (ex: "Conto Corrente Nubank") |
| `bank_name` | string | sim | Nome do banco de destino |
| `pix_key` | string \| null | não | Chave Pix (CPF, CNPJ, e-mail, telefone, aleatória) |
| `pix_titular` | string \| null | não | Nome do titular da chave Pix |
| `transfer_details` | object \| null | não | Dados bancários adicionais (agência, conta, etc.) |

**Request example:**

```json
{
  "name": "Conta Nubank PJ",
  "bank_name": "Nubank",
  "pix_key": "123.456.789-00",
  "pix_titular": "João da Silva",
  "transfer_details": {
    "agency": "0001",
    "account": "1234567-8"
  }
}
```

**Response 200:**

```json
{
  "success": true,
  "message": "conta bancária criada com sucesso",
  "conta": {
    "id": 1,
    "status": "active"
  }
}
```

---

### 2. Listar contas bancárias

```
GET /estabelecimentos/contasbancarias
Authorization: Bearer <token>
```

Retorna todas as contas bancárias da loja autenticada.

**Response 200:**

```json
{
  "success": true,
  "message": "contas bancárias listadas com sucesso",
  "contas": [
    {
      "id": 1,
      "shop_id": 3,
      "name": "Conta Nubank PJ",
      "bank_name": "Nubank",
      "pix": {
        "key": "123.456.789-00",
        "titular": "João da Silva"
      },
      "transfer_details": {
        "agency": "0001",
        "account": "1234567-8"
      },
      "created_at": "2026-05-25T14:30:00-03:00"
    }
  ]
}
```

---

### 3. Atualizar conta bancária

```
PUT /estabelecimentos/contasbancarias/:id
Authorization: Bearer <token>
```

**Body** (todos os campos opcionais — envia só os que deseja alterar):

| Campo | Tipo | Descrição |
|---|---|---|
| `name` | string | Novo apelido |
| `bank_name` | string | Novo nome do banco |
| `pix_key` | string \| null | Nova chave Pix (null para limpar) |
| `pix_titular` | string \| null | Novo titular (null para limpar) |
| `transfer_details` | object \| null | Novos dados bancários (null para limpar) |

**Request example:**

```json
{
  "name": "Nubank - Conta Corrente"
}
```

**Response 200:**

```json
{
  "success": true,
  "message": "conta bancária atualizada com sucesso"
}
```

**Response 404:**

```json
{
  "success": false,
  "message": "recurso não encontrado"
}
```

> ⚠️ A validação verifica se a conta bancária pertence à loja autenticada.  
> Retorna 404 se o `:id` não existir ou não pertencer à loja.

---

### 4. Remover conta bancária

```
DELETE /estabelecimentos/contasbancarias/:id
Authorization: Bearer <token>
```

**Response 200:**

```json
{
  "success": true,
  "message": "conta bancária removida com sucesso"
}
```

> ⚠️ A remoção verifica ownership da loja (retorna 404 caso contrário).  
> Saques que referenciam esta conta ficam com `bank_account_id = NULL` (ON DELETE SET NULL).

---

## Saques (Withdrawals)

### 1. Criar saque

```
POST /estabelecimentos/saques
Authorization: Bearer <token>
```

**Body:**

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `description` | string | sim | Descrição/motivo do saque |
| `asked_value` | string (decimal) | sim | Valor solicitado (ex: `"100.00"`) |
| `bank_account` | int | sim | ID da conta bancária previamente cadastrada |

**Request example:**

```json
{
  "description": "Saque semanal",
  "asked_value": "250.00",
  "bank_account": 1
}
```

**Response 200:**

```json
{
  "success": true,
  "message": "saque solicitado com sucesso",
  "saque": {
    "id": 3,
    "status": "pending"
  }
}
```

> ⚠️ O `bank_account` é validado contra as contas da própria loja.  
> Retorna 404 se a conta não existir ou não pertencer à loja.

O saque é criado com status `pending`.

---

### 2. Listar saques

```
GET /estabelecimentos/saques
Authorization: Bearer <token>
```

Retorna todos os saques da loja autenticada.

**Response 200:**

```json
{
  "success": true,
  "message": "saques listados com sucesso",
  "saques": [
    {
      "id": 1,
      "description": "Saque semanal",
      "asked_value": "250.00",
      "user_name": "João da Silva",
      "shop_name": "Loja Centro",
      "bank_name": "Nubank",
      "pix": {
        "key": "123.456.789-00",
        "titular": "João da Silva"
      },
      "transfer_details": {
        "agency": "0001",
        "account": "1234567-8"
      },
      "status": "pending",
      "created_at": "2026-05-25T14:30:00-03:00"
    }
  ]
}
```

> Os campos `pix` e `transfer_details` vêm da `bank_account` referenciada.  
> Se a conta foi removida (ON DELETE SET NULL), esses campos vêm `null`.

---

### 3. Visualizar saque

```
GET /estabelecimentos/saques/:id
Authorization: Bearer <token>
```

**Response 200:**

```json
{
  "success": true,
  "message": "saque encontrado com sucesso",
  "saque": {
    "id": 1,
    "description": "Saque semanal",
    "asked_value": "250.00",
    "user_name": "João da Silva",
    "shop_name": "Loja Centro",
    "bank_name": "Nubank",
    "pix": {
      "key": "123.456.789-00",
      "titular": "João da Silva"
    },
    "transfer_details": {
      "agency": "0001",
      "account": "1234567-8"
    },
    "status": "pending",
    "created_at": "2026-05-25T14:30:00-03:00"
  }
}
```

**Response 404:**

```json
{
  "success": false,
  "message": "saque não encontrado"
}
```

---

### 4. Atualizar status do saque

```
PUT /estabelecimentos/saques/:id
Authorization: Bearer <token>
```

**Body:**

| Campo | Tipo | Obrigatório | Valores aceitos |
|---|---|---|---|
| `status` | string | sim | `approved`, `rejected`, `cancelled` |

**Request example:**

```json
{
  "status": "approved"
}
```

**Response 200:**

```json
{
  "success": true,
  "message": "status do saque atualizado com sucesso"
}
```

---

### 5. Remover saque

```
DELETE /estabelecimentos/saques/:id
Authorization: Bearer <token>
```

**Response 200:**

```json
{
  "success": true,
  "message": "saque removido com sucesso"
}
```

> A remoção é física (hard delete).

---

## Modelo de dados

### `bank_accounts`

| Coluna | Tipo | Descrição |
|---|---|---|
| `id` | serial | PK |
| `shop_id` | int (FK → shops) | Loja dona da conta |
| `name` | varchar | Apelido da conta |
| `bank_name` | varchar | Nome do banco |
| `pix_key` | varchar \| null | Chave Pix |
| `pix_titular` | varchar \| null | Titular da chave |
| `transfer_details` | jsonb \| null | Agência, conta, etc. |
| `created_at` | timestamptz | Data de criação |

### `saques`

| Coluna | Tipo | Descrição |
|---|---|---|
| `id` | serial | PK |
| `shop_id` | int (FK → shops) | Loja solicitante |
| `user_id` | int (FK → users) | Usuário que solicitou |
| `description` | varchar | Motivo do saque |
| `asked_value` | numeric | Valor solicitado |
| `bank_account_id` | int (FK → bank_accounts) | **Conta bancária de destino** (pode ser NULL se a conta foi deletada) |
| `status` | varchar | `pending` \| `approved` \| `rejected` \| `cancelled` |
| `created_at` | timestamptz | Data de criação |

> Os campos `pix_key`, `pix_titular`, `transfer_details` e `bank_name` foram migrados para `bank_accounts`.  
> O saque agora referencia a conta via `bank_account_id` — os dados Pix e bancários são lidos da conta associada via JOIN.

---

## Regras

- Apenas lojas autenticadas (role `shop`) têm acesso a todas as rotas acima.
- `name`, `bank_name` na criação de conta bancária são obrigatórios.
- `description`, `asked_value`, `bank_account` na criação de saque são obrigatórios.
- O `bank_account` informado no saque é validado como pertencente à loja (404 se não).
- A remoção de conta bancária verifica ownership da loja (404 se não pertencer).
- Saques com conta bancária removida ficam com `bank_account_id = NULL` — os campos Pix/bancários aparecem vazios na resposta.
- Status aceitos para atualização: `approved`, `rejected`, `cancelled`.
- Remoção de saque é física (hard delete).
