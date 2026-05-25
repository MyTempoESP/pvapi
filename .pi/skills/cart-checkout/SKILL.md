---
name: cart-checkout
description: "Adds a product to the cart and checks out on pvauto.linhazero.com. Runs the full e-commerce flow: add to cart → retrieve package options → confirm checkout. Use when asked to add a product to cart, complete a purchase, or test the checkout flow against the pvauto API."
---

# Cart & Checkout

Automates the complete purchase flow against the pvauto API: add a product to the cart, fetch available pickup packages, and confirm checkout.

## Workflow

### 1. List Products (Optional)

```bash
curl -s https://pvauto.linhazero.com/produtos | python3 -m json.tool
```

### 2. Full Checkout Flow

The script `scripts/checkout.sh` runs the end-to-end flow:

```bash
SKILL_DIR="$(dirname "$(readlink -f "$0")")"
"$SKILL_DIR/scripts/checkout.sh"
```

**Default behavior**: uses product ID `1`, quantity `1`, auto-selects the first package with stock > 0 (displays all packages with stock info), and generates a unique session ID.

**Custom options**:

```bash
"$SKILL_DIR/scripts/checkout.sh" --product-id 1 --quantity 2 --package-id 4 --base-url https://pvauto.linhazero.com
```

**Parameters:**

| Flag | Default | Description |
|------|---------|-------------|
| `--product-id` | `1` | Product ID to add to cart |
| `--quantity` | `1` | Quantity to add |
| `--package-id` | auto (first in-stock) | Specific package ID to use for checkout |
| `--base-url` | `https://pvauto.linhazero.com` | API base URL |
| `--session-id` | auto-generated | Cart session identifier |
| `--value-paid` | derived from package | Amount paid (auto-derived from package price × quantity) |
| `--external-ref` | auto-generated | External payment reference |

### 3. Manual Steps (Step by Step)

If the script is unavailable, execute these API calls:

```bash
SESSION="cart_$(date +%s)_$$"
BASE="https://pvauto.linhazero.com"

# Add to cart
curl -s -X POST "$BASE/carrinho/adicionar" \
  -H "Content-Type: application/json" \
  -d "{\"product_id\":1,\"quantity\":1,\"session_id\":\"$SESSION\",\"personalization\":{}}"

# Get package options
curl -s "$BASE/carrinho/pacotes?session_id=$SESSION"

# Confirm checkout (fill package_id and value_paid from package options response)
curl -s -X POST "$BASE/checkouts/confirm?session_id=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{"value_paid":"6.00","external_reference":"EXT-REF-'"$(date +%s)"'","items":[{"package_id":3,"quantity":1}]}'
```

### API Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/produtos` | GET | List all products |
| `/carrinho/adicionar` | POST | Add product to cart (`product_id`, `quantity`, `session_id`, `personalization`) |
| `/carrinho/dados?session_id=` | GET | View cart contents |
| `/carrinho/pacotes?session_id=` | GET | List available pickup packages |
| `/carrinho/remover/:item?session_id=` | DELETE | Remove item from cart |
| `/carrinho/atualizar/:item?session_id=` | PUT | Update item quantity |
| `/checkouts/confirm?session_id=` | POST | Confirm checkout (`value_paid`, `external_reference`, `items[{package_id,quantity}]`) |

### Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 404 | Cart not found or already checked out |
| 409 | Insufficient stock |
| 422 | Empty cart or package not found |
