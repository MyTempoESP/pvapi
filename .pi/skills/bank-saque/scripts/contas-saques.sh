#!/usr/bin/env bash
set -euo pipefail

BASE_URL="https://pvauto.linhazero.com"
EMAIL=""
PASSWORD=""

NAME=""
BANK_NAME=""
PIX_KEY=""
PIX_TITULAR=""

DESCRIPTION=""
ASKED_VALUE=""
BANK_ACCOUNT_ID=""

ACTION=""

usage() {
  cat <<EOF
Usage: $0 [options]

Interactive mode (default):
  Runs with prompts for credentials and actions.

Non-interactive mode (all flags provided):
  --email EMAIL         Login email (required)
  --password PASSWORD   Login password (required)
  --name NAME           Bank account name (required for create-bank)
  --bank-name NAME      Bank name, e.g. Nubank, Itaú (required for create-bank)
  --pix-key KEY         PIX key (optional for create-bank)
  --pix-titular NAME    PIX account holder (optional for create-bank)
  --description TEXT    Saque description (required for create-saque)
  --asked-value AMOUNT  Saque value, e.g. 250.00 (required for create-saque)
  --bank-account-id ID  Bank account ID (required for create-saque)
  --action ACTION       One of: create-bank, create-saque, full
                        "full" does login → create bank → create saque
                        (default: full)
  --base-url URL        API base URL (default: $BASE_URL)
  -h, --help            Show this help
EOF
  exit 0
}

log()  { echo "  [$1] $2"; }
dump() { echo "$1" | python3 -m json.tool 2>/dev/null || echo "$1"; }

ask() {
  local prompt="$1"
  local var_name="$2"
  local val
  read -r -p "$prompt: " val
  eval "$var_name=\$val"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --email)           EMAIL="$2"; shift 2 ;;
    --password)        PASSWORD="$2"; shift 2 ;;
    --name)            NAME="$2"; shift 2 ;;
    --bank-name)       BANK_NAME="$2"; shift 2 ;;
    --pix-key)         PIX_KEY="$2"; shift 2 ;;
    --pix-titular)     PIX_TITULAR="$2"; shift 2 ;;
    --description)     DESCRIPTION="$2"; shift 2 ;;
    --asked-value)     ASKED_VALUE="$2"; shift 2 ;;
    --bank-account-id) BANK_ACCOUNT_ID="$2"; shift 2 ;;
    --action)          ACTION="$2"; shift 2 ;;
    --base-url)        BASE_URL="$2"; shift 2 ;;
    -h|--help)         usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

# ── Prompt for credentials if not provided ───────────────────────────────────
if [[ -z "$EMAIL" ]]; then
  ask "Enter your login email" EMAIL
fi
if [[ -z "$PASSWORD" ]]; then
  ask "Enter your password" PASSWORD
fi
if [[ -z "$ACTION" ]]; then
  ACTION="full"
fi

# ── Step 1: Login ────────────────────────────────────────────────────────────
log "AUTH" "Logging in as $EMAIL"

LOGIN_RESP=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESP" | python3 -c "
import sys, json
d = json.load(sys.stdin)
token = d.get('access_token', '')
if token:
    print(token)
    sys.exit(0)
else:
    sys.exit(1)
") || {
  echo "FAILED to login:"
  dump "$LOGIN_RESP"
  exit 1
}
log "OK"   "Authenticated successfully"

AUTH_HEADER="Authorization: Bearer $TOKEN"

# ── create-bank: Create bank account ─────────────────────────────────────────
create_bank() {
  if [[ -z "$NAME" ]]; then
    ask "Bank account name (e.g. Conta Principal)" NAME
  fi
  if [[ -z "$BANK_NAME" ]]; then
    ask "Bank name (e.g. Nubank, Itaú, Bradesco)" BANK_NAME
  fi
  if [[ -z "$PIX_KEY" ]]; then
    ask "PIX key (optional, press Enter to skip)" PIX_KEY
  fi
  if [[ -z "$PIX_TITULAR" ]]; then
    ask "PIX account holder (optional, press Enter to skip)" PIX_TITULAR
  fi

  local payload
  payload=$(python3 -c "
import json
p = {'name': '$NAME', 'bank_name': '$BANK_NAME'}
if '$PIX_KEY':   p['pix_key'] = '$PIX_KEY'
if '$PIX_TITULAR': p['pix_titular'] = '$PIX_TITULAR'
print(json.dumps(p))
")

  log "BANK" "Creating bank account: $NAME at $BANK_NAME"

  BANK_RESP=$(curl -s -X POST "$BASE_URL/estabelecimentos/contasbancarias" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "$payload")

  echo "$BANK_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
    echo "FAILED to create bank account:"
    dump "$BANK_RESP"
    exit 1
  }

  BANK_ACCOUNT_ID=$(echo "$BANK_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['conta']['id'])")
  BANK_ACCOUNT_NAME="$NAME"
  BANK_NAME_RESP="$BANK_NAME"

  echo ""
  echo "  ═══ Bank Account Created ═══"
  echo "  ID:        $BANK_ACCOUNT_ID"
  echo "  Name:      $BANK_ACCOUNT_NAME"
  echo "  Bank:      $BANK_NAME_RESP"
  echo ""
}

# ── create-saque: Create withdrawal ──────────────────────────────────────────
create_saque() {
  if [[ -z "$BANK_ACCOUNT_ID" ]]; then
    # List bank accounts for the user to pick from
    log "LIST" "Fetching your bank accounts"

    LIST_BA_RESP=$(curl -s "$BASE_URL/estabelecimentos/contasbancarias" \
      -H "$AUTH_HEADER")

    echo "$LIST_BA_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
      echo "FAILED to list bank accounts:"
      dump "$LIST_BA_RESP"
      exit 1
    }

    echo ""
    echo "  Your bank accounts:"
    echo "  ┌──────┬──────────────────────┬──────────┐"
    echo "  │  ID  │ Name                 │ Bank     │"
    echo "  ├──────┼──────────────────────┼──────────┤"
    echo "$LIST_BA_RESP" | python3 -c "
import sys, json
d = json.load(sys.stdin)
for c in d.get('contas', []):
    print(f\"  │ {c['id']:<4} │ {c['name']:<20} │ {c['bank_name']:<8} │\")
" 2>/dev/null || echo "  │ (none) │"
    echo "  └──────┴──────────────────────┴──────────┘"
    echo ""

    ask "Enter the bank account ID for this saque" BANK_ACCOUNT_ID
  fi

  if [[ -z "$DESCRIPTION" ]]; then
    ask "Saque description (e.g. Saque de teste)" DESCRIPTION
  fi
  if [[ -z "$ASKED_VALUE" ]]; then
    ask "Asked value (e.g. 250.00)" ASKED_VALUE
  fi

  log "SAQUE" "Creating saque: $DESCRIPTION — R\$ $ASKED_VALUE"

  SAQUE_RESP=$(curl -s -X POST "$BASE_URL/estabelecimentos/saques" \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d "{\"description\":\"$DESCRIPTION\",\"asked_value\":\"$ASKED_VALUE\",\"bank_account\":$BANK_ACCOUNT_ID}")

  echo "$SAQUE_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
    echo "FAILED to create saque:"
    dump "$SAQUE_RESP"
    exit 1
  }

  SAQUE_ID=$(echo "$SAQUE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['saque']['id'])")
  SAQUE_STATUS=$(echo "$SAQUE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['saque']['status'])")
  SAQUE_BANK="$BANK_NAME"

  echo ""
  echo "  ═══ Saque Created ═══"
  echo "  ID:          $SAQUE_ID"
  echo "  Description: $DESCRIPTION"
  echo "  Value:       R\$ $ASKED_VALUE"
  echo "  Bank:        $SAQUE_BANK"
  echo "  Status:      $SAQUE_STATUS"
  echo ""
}

# ── Execute requested action ─────────────────────────────────────────────────
case "$ACTION" in
  create-bank)
    create_bank
    ;;
  create-saque)
    create_saque
    ;;
  full)
    create_bank
    create_saque
    echo "══════════════════════════════════════════════"
    echo "  Bank account and saque created successfully."
    echo "══════════════════════════════════════════════"
    ;;
  *)
    echo "Unknown action: $ACTION"
    usage
    ;;
esac
