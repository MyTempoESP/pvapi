# Saques (Withdrawals)

Loja solicita retirada de saldo via `/estabelecimentos/saques`.  
Todas as rotas exigem autenticação com role `shop`.

---

## Endpoints

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
| `bank_name` | string | sim | Nome do banco de destino |
| `pix_cpf` | string \| null | não | CPF da chave Pix |
| `pix_titular` | string \| null | não | Nome do titular da conta |
| `transfer_details` | object \| null | não | Dados bancários adicionais (agência, conta, etc.) |

**Request example:**

```json
{
  "description": "Saque semanal",
  "asked_value": "250.00",
  "bank_name": "Nubank",
  "pix_cpf": "123.456.789-00",
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
  "message": "saque solicitado com sucesso",
  "saque": {
    "id": 3,
    "status": "pending"
  }
}
```

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
        "cpf": "123.456.789-00",
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
      "cpf": "123.456.789-00",
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

### 4. Atualizar status

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

---

## Fluxo completo

```
1. POST   /estabelecimentos/saques     ← criar saque (status: pending)
2. GET    /estabelecimentos/saques     ← listar todos
3. GET    /estabelecimentos/saques/:id ← ver detalhes de um saque
4. PUT    /estabelecimentos/saques/:id ← aprovar/rejeitar/cancelar
5. DELETE /estabelecimentos/saques/:id ← remover (qualquer status)
```

## Regras

- Apenas usuários com role `shop` têm acesso.
- `asked_value` e `bank_name` são obrigatórios na criação.
- `pix`, `transfer_details` são opcionais — retornam somente quando informados.
- Status aceitos para atualização: `approved`, `rejected`, `cancelled`.
- A remoção é física (hard delete).
