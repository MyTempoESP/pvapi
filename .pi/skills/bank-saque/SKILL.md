---
name: bank-saque
description: "Creates a bank account (conta bancária) and a withdrawal (saque) on pvauto.linhazero.com. Authenticates with credentials, creates the bank account, then creates a saque tied to it. Use when asked to create a bank account, create a withdrawal/saque, or run the bank/saque flow against the pvauto API."
---

# Bank Account & Saque

Automates the banking flow against the pvauto API: authenticate → create bank account → create withdrawal (saque).

**Credentials**: you must provide an email and password for a registered shop account on the API.

## Workflow

### 1. Full Flow (Recommended)

Runs login → create bank account → create saque, prompting for any missing info:

```bash
SKILL_DIR="$(dirname "$(readlink -f "$0")")"
"$SKILL_DIR/scripts/contas-saques.sh"
```

You will be prompted for:
1. **Email** – your shop login email
2. **Password** – your shop login password
3. **Bank account name** – e.g. `Conta Principal`
4. **Bank name** – e.g. `Nubank`, `Itaú`, `Bradesco`
5. **PIX key** (optional) – PIX identifier
6. **PIX account holder** (optional) – name on the PIX account
7. **Bank account ID** (if no bank accounts exist yet, it uses the newly created one; otherwise lists existing accounts and asks you to pick)
8. **Saque description** – e.g. `Saque de teste`
9. **Asked value** – e.g. `250.00`

### 2. Non-Interactive Mode

All arguments provided via flags:

```bash
"$SKILL_DIR/scripts/contas-saques.sh" \
  --email loja@example.com \
  --password minha-senha \
  --name "Conta Principal" \
  --bank-name "Nubank" \
  --pix-key "12345678900" \
  --pix-titular "Loja Teste" \
  --description "Saque de teste" \
  --asked-value "250.00" \
  --action full
```

### 3. Bank Account Only

```bash
"$SKILL_DIR/scripts/contas-saques.sh" \
  --email loja@example.com \
  --password minha-senha \
  --name "Conta Principal" \
  --bank-name "Nubank" \
  --action create-bank
```

### 4. Saque Only (using an existing bank account)

```bash
"$SKILL_DIR/scripts/contas-saques.sh" \
  --email loja@example.com \
  --password minha-senha \
  --bank-account-id 1 \
  --description "Saque de teste" \
  --asked-value "250.00" \
  --action create-saque
```

If `--bank-account-id` is omitted when creating a saque, the script lists your existing bank accounts and lets you pick one.

### Parameters

| Flag | Default | Description |
|------|---------|-------------|
| `--email` | *prompted* | Login email (required) |
| `--password` | *prompted* | Login password (required) |
| `--name` | *prompted* | Bank account name |
| `--bank-name` | *prompted* | Bank name (Nubank, Itaú, Bradesco, etc.) |
| `--pix-key` | *prompted (optional)* | PIX key for the account |
| `--pix-titular` | *prompted (optional)* | PIX account holder name |
| `--description` | *prompted* | Saque description |
| `--asked-value` | *prompted* | Saque value (e.g. 250.00) |
| `--bank-account-id` | *auto / prompted* | Existing bank account ID |
| `--action` | `full` | One of: `full`, `create-bank`, `create-saque` |
| `--base-url` | `https://pvauto.linhazero.com` | API base URL |
| `-h`, `--help` | | Show help |

## Manual Steps (Step by Step)

If the script is unavailable, execute these API calls:

```bash
BASE="https://pvauto.linhazero.com"

# 1. Login — save the token
TOKEN=$(curl -s -X POST "$BASE/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"YOUR_EMAIL","password":"YOUR_PASSWORD"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

AUTH="Authorization: Bearer $TOKEN"

# 2. Create bank account
curl -s -X POST "$BASE/estabelecimentos/contasbancarias" \
  -H "Content-Type: application/json" \
  -H "$AUTH" \
  -d '{
    "name": "Conta Principal",
    "bank_name": "Nubank",
    "pix_key": "12345678900",
    "pix_titular": "Loja Teste"
  }'

# 3. List bank accounts (to get the ID)
curl -s "$BASE/estabelecimentos/contasbancarias" -H "$AUTH"

# 4. Create saque (use the bank account ID from step 2 or 3)
curl -s -X POST "$BASE/estabelecimentos/saques" \
  -H "Content-Type: application/json" \
  -H "$AUTH" \
  -d '{
    "description": "Saque de teste",
    "asked_value": "250.00",
    "bank_account": 1
  }'
```

## API Reference

### Bank Accounts

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/estabelecimentos/contasbancarias` | GET | List all bank accounts for the authenticated shop |
| `/estabelecimentos/contasbancarias` | POST | Create bank account (`name`, `bank_name`, `pix_key?`, `pix_titular?`) |
| `/estabelecimentos/contasbancarias/:id` | PUT | Update bank account |
| `/estabelecimentos/contasbancarias/:id` | DELETE | Delete bank account |

### Saques (Withdrawals)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/estabelecimentos/saques` | GET | List all saques for the authenticated shop |
| `/estabelecimentos/saques` | POST | Create saque (`description`, `asked_value`, `bank_account`) |
| `/estabelecimentos/saques/:id` | GET | View saque details |
| `/estabelecimentos/saques/:id` | PUT | Update saque status (`status`: `pending`, `approved`, `cancelled`) |
| `/estabelecimentos/saques/:id` | DELETE | Delete saque |

### Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/login` | POST | Authenticate (`email`, `password`) → returns JWT `token` |
| `/register` | POST | Register new account |

### Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request (missing/invalid fields) |
| 401 | Unauthorized (missing or invalid token) |
| 404 | Not found |
