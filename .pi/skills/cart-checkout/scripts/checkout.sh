#!/usr/bin/env bash
set -euo pipefail

PRODUCT_ID=1
QUANTITY=1
BASE_URL="https://pvauto.linhazero.com"
PACKAGE_ID=""

usage() {
  cat <<EOF
Usage: $0 [options]

Options:
  --product-id N      Product ID (default: $PRODUCT_ID)
  --quantity N        Quantity (default: $QUANTITY)
  --package-id N      Specific package to use (default: auto-select first with stock > 0)
  --base-url URL      API base URL (default: $BASE_URL)
  --session-id ID     Cart session ID (default: auto-generated)
  --value-paid AMOUNT Override value_paid (default: derived from package price)
  --external-ref REF  External payment reference (default: auto-generated)
  -h, --help          Show this help
EOF
  exit 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --product-id)   PRODUCT_ID="$2"; shift 2 ;;
    --quantity)     QUANTITY="$2"; shift 2 ;;
    --package-id)   PACKAGE_ID="$2"; shift 2 ;;
    --base-url)     BASE_URL="$2"; shift 2 ;;
    --session-id)   SESSION_ID="$2"; shift 2 ;;
    --value-paid)   VALUE_PAID_OVERRIDE="$2"; shift 2 ;;
    --external-ref) EXTERNAL_REF="$2"; shift 2 ;;
    -h|--help)      usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

SESSION_ID="${SESSION_ID:-cart_$(date +%s)_$$}"
EXTERNAL_REF="${EXTERNAL_REF:-EXT-REF-$(date +%s)}"

log()  { echo "  [$1] $2"; }
dump() { echo "$1" | python3 -m json.tool 2>/dev/null || echo "$1"; }

# ── Step 1: Add product to cart ──────────────────────────────────────────────
log "CART" "Adding product_id=$PRODUCT_ID qty=$QUANTITY"

ADD_RESP=$(curl -s -X POST "$BASE_URL/carrinho/adicionar" \
  -H "Content-Type: application/json" \
  -d "{\"product_id\":$PRODUCT_ID,\"quantity\":$QUANTITY,\"session_id\":\"$SESSION_ID\",\"personalization\":{}}")

echo "$ADD_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
  echo "FAILED to add product:"
  dump "$ADD_RESP"
  exit 1
}
log "OK"   "Product added"

# ── Step 2: Get package options ──────────────────────────────────────────────
log "PKG"  "Fetching package options"

PKG_RESP=$(curl -s "$BASE_URL/carrinho/pacotes?session_id=$SESSION_ID")

ALL_PACKAGES=$(echo "$PKG_RESP" | python3 -c "
import sys, json
d = json.load(sys.stdin)
products = d.get('products', [])
if not products:
    sys.exit(1)
pkgs = products[0].get('packages', [])
if not pkgs:
    sys.exit(1)
for p in pkgs:
    print(f\"{p['package_id']}|{p['price']}|{p['stock']}|{p['pickup']['name']}\")
") || {
  echo "FAILED: no packages available:"
  dump "$PKG_RESP"
  exit 1
}

# Display available packages
echo ""
echo "  Available packages for product #$PRODUCT_ID:"
echo "  ┌────┬─────────┬───────┬──────────┐"
echo "  │ ID │  Price   │ Stock │ Pickup   │"
echo "  ├────┼─────────┼───────┼──────────┤"
while IFS='|' read -r pid price stock pickup; do
  printf "  │ %-2s │ R\$%5s │  %-4s │ %-8s │\n" "$pid" "$price" "$stock" "$pickup"
done <<< "$ALL_PACKAGES"
echo "  └────┴─────────┴───────┴──────────┘"
echo ""

# Select package
if [[ -n "$PACKAGE_ID" ]]; then
  # User specified a package — validate it exists and has stock
  SELECTED=$(echo "$ALL_PACKAGES" | awk -F'|' -v pid="$PACKAGE_ID" '$1 == pid')
  if [[ -z "$SELECTED" ]]; then
    echo "ERROR: package #$PACKAGE_ID not found"
    exit 1
  fi
  pkg_stock=$(echo "$SELECTED" | cut -d'|' -f3)
  if [[ "$pkg_stock" -eq 0 ]]; then
    echo "ERROR: package #$PACKAGE_ID has no stock"
    exit 1
  fi
else
  # Auto-select first package with stock > 0
  SELECTED=$(echo "$ALL_PACKAGES" | awk -F'|' '$3 > 0 {print; exit}')
  if [[ -z "$SELECTED" ]]; then
    echo "ERROR: no packages with available stock"
    exit 1
  fi
fi

PACKAGE_ID=$(echo "$SELECTED"    | cut -d'|' -f1)
PACKAGE_PRICE=$(echo "$SELECTED" | cut -d'|' -f2)
PICKUP_NAME=$(echo "$SELECTED"   | cut -d'|' -f4)

log "OK"   "Using package #$PACKAGE_ID — price: $PACKAGE_PRICE — pickup: $PICKUP_NAME"

# ── Step 3: Confirm checkout ─────────────────────────────────────────────────
if [[ -n "${VALUE_PAID_OVERRIDE:-}" ]]; then
  VALUE_PAID="$VALUE_PAID_OVERRIDE"
else
  VALUE_PAID=$(python3 -c "print(f'{$PACKAGE_PRICE * $QUANTITY:.2f}')")
fi

log "CHK"  "Confirming checkout — value_paid=$VALUE_PAID — ref=$EXTERNAL_REF"

CHECKOUT_RESP=$(curl -s -X POST "$BASE_URL/checkouts/confirm?session_id=$SESSION_ID" \
  -H "Content-Type: application/json" \
  -d "{\"value_paid\":\"$VALUE_PAID\",\"external_reference\":\"$EXTERNAL_REF\",\"items\":[{\"package_id\":$PACKAGE_ID,\"quantity\":$QUANTITY}]}")

echo "$CHECKOUT_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); sys.exit(0 if d.get('success') else 1)" || {
  echo "FAILED checkout:"
  dump "$CHECKOUT_RESP"
  exit 1
}

UNIQUE_ID=$(echo "$CHECKOUT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['unique_id'])")
STATUS=$(echo "$CHECKOUT_RESP"    | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])")

echo ""
echo "═══ Checkout Confirmed ═══"
echo "  Session:     $SESSION_ID"
echo "  Product:     #$PRODUCT_ID × $QUANTITY"
echo "  Package:     #$PACKAGE_ID"
echo "  Value paid:  R\$ $VALUE_PAID"
echo "  Reference:   $EXTERNAL_REF"
echo "  Unique ID:   $UNIQUE_ID"
echo "  Status:      $STATUS"
