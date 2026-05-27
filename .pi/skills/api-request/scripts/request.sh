#!/usr/bin/env bash
set -euo pipefail

BASE_URL="https://pvauto.linhazero.com"
METHOD="GET"
ENDPOINT=""
DATA=""
DATA_FILE=""
QUERY=""
AUTH=false
EMAIL=""
PASSWORD=""
VERBOSE=false
RAW=false

usage() {
  cat <<EOF
Usage: $0 [options]

Generic API request runner for $BASE_URL.

Options:
  --method METHOD       HTTP method: GET, POST, PUT, DELETE, PATCH (default: GET)
  --endpoint PATH       API path, e.g. /produtos, /login
  --data JSON           Request body as inline JSON string
  --data-file PATH      Read request body from a file
  --query STRING        Query string, e.g. "key1=val1&key2=val2"
  --auth                Authenticate and attach Bearer token
  --email EMAIL         Login email (required when --auth is set)
  --password PASSWORD   Login password (required when --auth is set)
  --base-url URL        API base URL (default: $BASE_URL)
  --verbose             Print request details (method, headers, body)
  --raw                 Print raw JSON (skip pretty-print)
  -h, --help            Show this help
EOF
  exit 0
}

log()  { echo "  [$1] $2"; }
dump() {
  if $RAW; then
    echo "$1"
  else
    echo "$1" | python3 -m json.tool 2>/dev/null || echo "$1"
  fi
}

ask() {
  local prompt="$1"
  local var_name="$2"
  local val
  read -r -p "$prompt: " val
  eval "$var_name=\$val"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --method)     METHOD="$2"; shift 2 ;;
    --endpoint)   ENDPOINT="$2"; shift 2 ;;
    --data)       DATA="$2"; shift 2 ;;
    --data-file)  DATA_FILE="$2"; shift 2 ;;
    --query)      QUERY="$2"; shift 2 ;;
    --auth)       AUTH=true; shift ;;
    --email)      EMAIL="$2"; shift 2 ;;
    --password)   PASSWORD="$2"; shift 2 ;;
    --base-url)   BASE_URL="$2"; shift 2 ;;
    --verbose)    VERBOSE=true; shift ;;
    --raw)        RAW=true; shift ;;
    -h|--help)    usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

# ── Interactive mode: prompt for missing values ──────────────────────────────
if [[ -z "$ENDPOINT" ]]; then
  echo "Interactive mode — fill in the request details."
  echo ""

  ask "HTTP method [GET/POST/PUT/DELETE/PATCH]" METHOD
  METHOD=${METHOD:-GET}
  ask "Endpoint (e.g. /produtos)" ENDPOINT

  read -r -p "Authenticate? [y/N]: " yn
  if [[ "$yn" =~ ^[Yy] ]]; then
    AUTH=true
  fi

  if [[ "$METHOD" =~ ^(POST|PUT|PATCH)$ ]]; then
    read -r -p "JSON body (inline, Enter to skip): " DATA
    if [[ -z "$DATA" ]]; then
      read -r -p "Body from file (Enter to skip): " DATA_FILE
    fi
  fi

  read -r -p "Query string (Enter to skip): " QUERY

  echo ""
fi

# ── Validate ─────────────────────────────────────────────────────────────────
METHOD=$(echo "$METHOD" | tr '[:lower:]' '[:upper:]')
case "$METHOD" in
  GET|POST|PUT|DELETE|PATCH) ;;
  *) echo "ERROR: invalid HTTP method: $METHOD"; exit 1 ;;
esac

# ── Authenticate if requested ────────────────────────────────────────────────
AUTH_HEADER=""
if $AUTH; then
  if [[ -z "$EMAIL" ]]; then
    ask "Login email" EMAIL
  fi
  if [[ -z "$PASSWORD" ]]; then
    ask "Login password" PASSWORD
  fi

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
  log "OK"   "Authenticated"
  AUTH_HEADER="-H \"Authorization: Bearer $TOKEN\""
fi

# ── Build body ───────────────────────────────────────────────────────────────
BODY_ARG=""
if [[ -n "$DATA_FILE" ]]; then
  if [[ ! -f "$DATA_FILE" ]]; then
    echo "ERROR: data file not found: $DATA_FILE"
    exit 1
  fi
  DATA=$(cat "$DATA_FILE")
fi

if [[ -n "$DATA" ]]; then
  # Validate JSON
  echo "$DATA" | python3 -m json.tool >/dev/null 2>&1 || {
    echo "ERROR: invalid JSON body"
    echo "$DATA" | python3 -m json.tool 2>&1 || true
    exit 1
  }
  BODY_ARG="-d '$DATA'"
fi

# ── Build URL ────────────────────────────────────────────────────────────────
URL="$BASE_URL$ENDPOINT"
if [[ -n "$QUERY" ]]; then
  URL="$URL?$QUERY"
fi

# ── Verbose output ───────────────────────────────────────────────────────────
if $VERBOSE; then
  echo ""
  echo "═══ Request ═══"
  echo "  Method:   $METHOD"
  echo "  URL:      $URL"
  if [[ -n "$AUTH_HEADER" ]]; then
    echo "  Auth:     Bearer ***"
  fi
  if [[ -n "$DATA" ]]; then
    echo "  Body:"
    echo "$DATA" | python3 -m json.tool 2>/dev/null | sed 's/^/    /'
  fi
  echo ""
fi

# ── Execute ──────────────────────────────────────────────────────────────────
log "REQ"  "$METHOD $URL"

# Build curl command as an array for safe quoting
CURL_ARGS=(-s -X "$METHOD")
CURL_ARGS+=(-H "Content-Type: application/json")

if [[ -n "$AUTH_HEADER" ]]; then
  CURL_ARGS+=(-H "Authorization: Bearer $TOKEN")
fi

if [[ -n "$DATA" ]]; then
  CURL_ARGS+=(-d "$DATA")
fi

CURL_ARGS+=("$URL")

RESP=$(curl "${CURL_ARGS[@]}")

# ── Output ───────────────────────────────────────────────────────────────────
echo ""
dump "$RESP"

# ── Status check ─────────────────────────────────────────────────────────────
HTTP_OK=$(echo "$RESP" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    ok = d.get('success', 'N/A')
    if ok is True:
        print('OK')
    else:
        msg = d.get('message', d.get('error', str(d)))
        print(f'FAIL: {msg}')
except:
    print('UNKNOWN')
" 2>/dev/null || echo "UNKNOWN")

if [[ "$HTTP_OK" == "OK" ]]; then
  echo ""
  log "OK" "Request succeeded"
elif [[ "$HTTP_OK" == UNKNOWN* ]]; then
  echo ""
  log "???" "Could not parse response"
else
  echo ""
  log "ERR" "$HTTP_OK"
  exit 1
fi
