#!/usr/bin/env bash
set -euo pipefail

BASE_URL="https://pvauto.linhazero.com"
TOKEN="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzk3OTE2NDYsImlkIjo1LCJvcmlnX2lhdCI6MTc3OTcwNTI0NiwicmlkIjoyLCJyb2xlIjoic2hvcCJ9.3waKJEm039rxnukgjOLd5eqdm6kCIxMQc8JTFFnZPU0"

log()  { echo "  [$1] $2"; }
dump() { echo "$1" | python3 -m json.tool 2>/dev/null || echo "$1"; }

# ── Step 1: POST /estabelecimentos/saques (Create) ────────────────────────────
log "CREATE" "Creating saque"

CREATE_RESP=$(curl -s -X POST "$BASE_URL/estabelecimentos/saques" \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{
    "description": "Teste de saque via script",
    "asked_value": "100.00",
    "bank_name": "Banco do Brasil",
    "pix_cpf": "123.456.789-00",
    "pix_titular": "Fulano de Tal",
    "transfer_details": {"agency": "1234", "account": "56789-0"}
  }')

echo "$CREATE_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
  echo "FAILED to create saque:"
  dump "$CREATE_RESP"
  exit 1
}

SAQUE_ID=$(echo "$CREATE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['saque']['id'])")
log "OK"   "Saque created — id=$SAQUE_ID"

# ── Step 2: GET /estabelecimentos/saques (List) ───────────────────────────────
log "LIST"  "Listing all saques"

LIST_RESP=$(curl -s "$BASE_URL/estabelecimentos/saques" \
  -H "Authorization: $TOKEN")

echo "$LIST_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
  echo "FAILED to list saques:"
  dump "$LIST_RESP"
  exit 1
}

COUNT=$(echo "$LIST_RESP" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['saques']))")
log "OK"   "Listed $COUNT saque(s)"

# ── Step 3: GET /estabelecimentos/saques/:saque (View) ────────────────────────
log "VIEW"  "Viewing saque #$SAQUE_ID"

VIEW_RESP=$(curl -s "$BASE_URL/estabelecimentos/saques/$SAQUE_ID" \
  -H "Authorization: $TOKEN")

echo "$VIEW_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
  echo "FAILED to view saque:"
  dump "$VIEW_RESP"
  exit 1
}

SAQUE_STATUS=$(echo "$VIEW_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['saque']['status'])")
log "OK"   "Saque #$SAQUE_ID — status=$SAQUE_STATUS"

# ── Step 4: PUT /estabelecimentos/saques/:saque (Update status → approved) ────
log "UPDATE" "Approving saque #$SAQUE_ID"

UPDATE_RESP=$(curl -s -X PUT "$BASE_URL/estabelecimentos/saques/$SAQUE_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{"status": "approved"}')

echo "$UPDATE_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
  echo "FAILED to update saque status:"
  dump "$UPDATE_RESP"
  exit 1
}

log "OK"   "Saque #$SAQUE_ID status updated to approved"

# ── Step 5: DELETE /estabelecimentos/saques/:saque ────────────────────────────
log "DELETE" "Deleting saque #$SAQUE_ID"

DELETE_RESP=$(curl -s -X DELETE "$BASE_URL/estabelecimentos/saques/$SAQUE_ID" \
  -H "Authorization: $TOKEN")

echo "$DELETE_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
  echo "FAILED to delete saque:"
  dump "$DELETE_RESP"
  exit 1
}

log "OK"   "Saque #$SAQUE_ID deleted"

echo ""
echo "═══════════════════════════════════════════"
echo "  All saque endpoint tests passed."
echo "═══════════════════════════════════════════"
