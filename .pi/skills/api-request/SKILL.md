---
name: api-request
description: "Generic API request runner for pvauto.linhazero.com. Sends arbitrary HTTP requests with optional authentication. Supports GET, POST, PUT, DELETE, PATCH with JSON body, query params, and both interactive and non-interactive modes. Use when asked to call any pvauto API endpoint not covered by specialized skills."
---

# API Request

Generic HTTP client for the pvauto API. Authenticates on demand, sends requests, and pretty-prints responses.

## Workflow

### 1. Interactive Mode

Prompts for method, endpoint, body, and auth:

```bash
SKILL_DIR="$(dirname "$(readlink -f "$0")")"
"$SKILL_DIR/scripts/request.sh"
```

### 2. Non-Interactive (GET)

```bash
"$SKILL_DIR/scripts/request.sh" \
  --method GET \
  --endpoint /produtos
```

### 3. Authenticated Request (POST with body)

```bash
"$SKILL_DIR/scripts/request.sh" \
  --method POST \
  --endpoint /estabelecimentos/contasbancarias \
  --auth \
  --email loja@example.com \
  --password minha-senha \
  --data '{"name":"Conta Principal","bank_name":"Nubank"}'
```

### 4. Authenticated GET with Query Params

```bash
"$SKILL_DIR/scripts/request.sh" \
  --method GET \
  --endpoint /carrinho/pacotes \
  --query "session_id=cart_123"
```

### 5. PUT / DELETE

```bash
"$SKILL_DIR/scripts/request.sh" \
  --method PUT \
  --endpoint /estabelecimentos/contasbancarias/5 \
  --auth \
  --email loja@example.com \
  --password minha-senha \
  --data '{"name":"Conta Atualizada"}'
```

### 6. Read Body from File

```bash
"$SKILL_DIR/scripts/request.sh" \
  --method POST \
  --endpoint /estabelecimentos/saques \
  --auth \
  --email loja@example.com \
  --password minha-senha \
  --data-file /tmp/saque.json
```

### Parameters

| Flag | Default | Description |
|------|---------|-------------|
| `--method` | `GET` | HTTP method: GET, POST, PUT, DELETE, PATCH |
| `--endpoint` | *prompted* | API path (e.g. `/produtos`, `/login`) |
| `--data` | — | JSON body as inline string |
| `--data-file` | — | JSON body read from a file |
| `--query` | — | Query string (e.g. `key1=val1&key2=val2`) |
| `--auth` | `false` | Authenticate and attach `Authorization: Bearer` header |
| `--email` | *prompted* | Login email (required when `--auth` is set) |
| `--password` | *prompted* | Login password (required when `--auth` is set) |
| `--base-url` | `https://pvauto.linhazero.com` | API base URL |
| `--verbose` | `false` | Print full request details (method, headers, body) |
| `--raw` | `false` | Print raw JSON (skip pretty-print) |
| `-h`, `--help` | | Show help |

## API Reference

The full pvauto API surface is documented across the specialized skills:

| Skill | Domain |
|-------|--------|
| `bank-saque` | `/login`, `/estabelecimentos/contasbancarias`, `/estabelecimentos/saques` |

This `api-request` skill wraps all of them generically. Use it for ad-hoc exploration, one-off calls, or endpoints not yet covered by a specialized skill.

### Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request (missing/invalid fields) |
| 401 | Unauthorized (missing or invalid token) |
| 404 | Not found |
| 409 | Conflict (e.g. insufficient stock) |
| 422 | Unprocessable entity |
