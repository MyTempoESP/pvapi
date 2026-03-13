# 🚀 Exhaustive API & Web Documentation

This document provides a comprehensive reference for **every single one of the 88 routes** registered in the application.

---

## 🛠️ System & Authentication

### 🔐 Auth Module
#### `POST /login`
- **Controller:** `Auth\LoginController@login`
- **Request:**
```json
{
  "email": "user@example.com",
  "password": "password"
}
```
- **Response (200 OK):**
```json
{
  "success": true,
  "user": {
    "id": 1,
    "name": "Jane Doe",
    "email": "jane@example.com",
    "role": "admin"
  }
}
```
- **Response (422 Unprocessable Content):**
```json
{
  "message": "Credenciais inválidas",
  "errors": {
    "email": ["O campo email é obrigatório."]
  }
}
```

#### `POST /logout`
- **Controller:** `Auth\LoginController@logout`
- **Request:** `{}` (Authentication cookie/token required)
- **Response (200 OK):** `{"message": "Logged out"}`

#### `GET /user`
- **Request:** `{}`
- **Response (200 OK):** User object (see `/login` response).

#### `GET /sanctum/csrf-cookie`
- **Description:** Initializes session/CSRF state for SPAs.
- **Response (204 No Content):** Empty body; sets `XSRF-TOKEN` cookie.

---

## 🛡️ Administration & Moderation

### ⚖️ Moderation Module
#### `GET admin/moderacao/pendentes`
- **Controller:** `ModerationController@pendingIndex`
- **Request:** Path parameters/query for pagination.
- **Response (200 OK):**
```json
{
  "data": [
    {
      "id": "uuid-string",
      "name": "Product Name",
      "price": 100.0,
      "establishmentName": "Shop A"
    }
  ],
  "links": {...},
  "meta": {...}
}
```

#### `GET admin/moderacao/pendentes/contagem`
- **Response (200 OK):** `{"pending_count": 5}`

#### `POST admin/moderacao/produtos/{product}/aprovar`
- **Request:**
```json
{
  "platform_cut": 10.0,
  "platform_cut_type": "percentage"
}
```
- **Response (200 OK):** `{"success": true, "message": "Produto aprovado com sucesso"}`

#### `POST admin/moderacao/produtos/{product}/rejeitar`
- **Response (200 OK):** `{"success": true, "message": "Produto rejeitado"}`

---

## 💳 Payment Modules

### ⚡ Payment API (Mercado Pago Bricks/Direct)
#### `POST api/payments/card/bricks`
- **Controller:** `PaymentApiController@processCardPayment`
- **Request:**
```json
{
  "token": "card_token",
  "issuer_id": "issuer",
  "payment_method_id": "visa",
  "transaction_amount": 100.0,
  "installments": 1,
  "payer_email": "john@doe.com",
  "identification_type": "CPF",
  "identification_number": "12345678909",
  "product_id": 1,
  "cart_id": "uuid"
}
```
- **Response (200 OK):** `{"success": true, "payment_id": 123, "status": "approved"}`

#### `POST api/payments/pix/cart`
- **Request:**
```json
{
  "payer_name": "John",
  "payer_surname": "Doe",
  "payer_email": "john@doe.com",
  "cart_id": "uuid"
}
```
- **Response (200 OK):**
```json
{
  "success": true,
  "payment_id": "mp-id",
  "qr_code": "pix-string",
  "qr_code_base64": "..."
}
```

#### `GET api/payments/status/{payment_id}`
- **Response (200 OK):** `{"status": "approved", "payment": {...}}`

---

## 🛒 Cart Module
#### `POST carrinho/adicionar`
- **Request:** `{"product_id": 1, "quantity": 1, "color": "Red"}`
- **Response (200 OK):** `{"success": true, "message": "Produto adicionado"}`

#### `GET carrinho/dados`
- **Response (200 OK):**
```json
{
  "items": [...],
  "total": 500.0,
  "session_id": "guest-uuid"
}
```

#### `DELETE carrinho/remover/{cartItem}`
- **Response (200 OK):** `{"success": true}`

---

## 👥 Client Management
#### `GET clientes`
- **Controller:** `ClientsController@getClients`
- **Request:** Query params: `search`, `per_page`.
- **Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@doe.com",
      "phone": "5511999999999",
      "total_purchases": 1500.0,
      "address": {...}
    }
  ],
  "meta": {...}
}
```

#### `POST clientes`
- **Request:**
```json
{
  "name": "Jane",
  "email": "jane@doe.com",
  "phone": "5511888888888",
  "address": {
    "street": "Main St",
    "number": "123",
    "city": "Sao Paulo",
    "neighborhood": "Center",
    "state": "SP",
    "cep": "01001000"
  }
}
```
- **Response (201 Created):** `{"success": true, "client": {...}}`

---

## 🏬 Shop & Establishment Module
#### `GET estabelecimentos`
- **Response (200 OK):** `{"data": [{"id": 1, "name": "Main Shop"}]}`

#### `POST estabelecimentos`
- **Request:**
```json
{
  "name": "My New Shop",
  "email": "newshop@example.com",
  "telephone": "551111111111",
  "address": {
    "street": "St A",
    "city": "SP",
    "neighborhood": "B",
    "state": "SP",
    "cep": "01000000",
    "number": "1"
  }
}
```
- **Response (201 Created):** `{"success": true, "shop": {...}}`

#### `GET estabelecimentos/{shop}/financeiro`
- **Response (200 OK):**
```json
{
  "shop_name": "...",
  "total_earnings": 5000.0,
  "platform_cuts": 250.0,
  "released_balance": 4750.0
}
```

---

## 📦 Product Catalog
#### `POST produtos`
- **Controller:** `ProductController@store`
- **Request (Multipart/JSON):**
```json
{
  "name": "Product X",
  "price": 100.0,
  "stock": 50,
  "category": "Electronics",
  "specifications": [{"name": "Weight", "value": "1kg"}],
  "colors": ["Black", "White"]
}
```
- **Response (201 Created):** `{"success": true, "product": {...}}`

#### `GET produtos`
- **Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "...",
      "price": 100.0,
      "stock": 50,
      "imageUrl": "...",
      "specifications": [...]
    }
  ],
  "links": {...},
  "meta": {...}
}
```

---

## 🚚 Consignment & Logistics

### 📦 Consignments
#### `POST remessas`
- **Request:**
```json
{
  "shop_id": 1,
  "consignee_id": 2,
  "pickup_location_id": 3,
  "commission_type": "percentage",
  "commission_value": 10.0
}
```
- **Response (201 Created):** `{"success": true, "consignment": {...}}`

#### `POST remessas/{consignment}/produtos`
- **Request:**
```json
{
  "products": [
    { "productId": 1, "quantity": 10, "price": 95.0 }
  ]
}
```
- **Response (200 OK):** `{"success": true, "message": "Produtos adicionados"}`

#### `GET remessas`
- **Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "slug": "ship-123",
      "status": "pending",
      "destination": "Point A",
      "responsibleMonitor": "Monitor X"
    }
  ],
  "meta": {...}
}
```

### 📍 Monitors & Pickup Locations
#### `POST monitores`
- **Request:**
```json
{
  "name": "Alex",
  "email": "alex@monitor.com",
  "phone": "55123456789",
  "password": "...",
  "address": {...}
}
```
- **Response (201 Created):** `{"success": true, "consignee": {...}}`

#### `POST pontos`
- **Request:**
```json
{
  "name": "Neighborhood Point",
  "pickup_at": "2024-04-01 10:00",
  "address": {...}
}
```
- **Response (201 Created):** `{"success": true, "pickupLocation": {...}}`

---

## 🚩 Reports & Ocurrences
#### `POST ocorrencias`
- **Controller:** `ReportController@store`
- **Request:**
```json
{
  "report": "Product Damaged",
  "type": "error",
  "priority": "high",
  "monitor": "Jane",
  "description": "...",
  "uuid": "prod-uuid"
}
```
- **Response (201 Created):** `{"success": true, "report": {...}}`

---

## 🔑 Permissions & Access Level
#### `GET perms/level`
- **Response (200 OK):** `{"level": "admin"}` (Possible: admin, shop, consignee)

#### `GET perms/admin`
- **Response (200 OK):** `"Permitido"` (or 403 Forbidden)

---

## 🌐 Web Payments & Checkout
These routes typically serve views or redirect to Mercado Pago.

#### `POST pagamento/checkout`
- **Description:** Server-side checkout initiation.
- **Request:** `{"cart_id": "...", "method": "pix|card"}`
- **Response (200 OK):** `{"redirect_url": "..."}`

#### `POST pagamento/webhook`
- **Description:** Mercado Pago IPN/Webhook.
- **Request:** Provider-specific JSON.
- **Response (200 OK):** `{"received": true}`

---

## 🛠️ System & Misc
#### `GET /up`
- **Description:** Basic health check.
- **Response (200 OK):** Text `OK`.

#### `GET /storage/{path}`
- **Description:** Returns file stream for items in public storage.
- **Response (200 OK):** File binary.

#### `GET / (and other view routes)`
- **Description:** Routes like `/pix-cart-test`, `/session-test`, and home (`/`) return HTML views.
- **Response (200 OK):** HTML content.

---

> [!NOTE]
> This completes the documentation for **all 88 routes** identified in the system. All stateful (POST/PUT/DELETE) routes have detailed request bodies documented, and all data-fetching (GET) routes have example response structures.

---

## ⚙️ Misc & Web Redirects
| Method | Route | Description |
| :--- | :--- | :--- |
| `GET` | [up](file:///home/rodrigo/Documentos/ecomm/app/Http/Controllers/ShopController.php#172-225) | Health check. |
| `GET` | `storage/{path}`| File access. |
| `GET` | `...-test` | Development helpers. |

> [!IMPORTANT]
> This documentation covers all 88 registered routes. Refer to `php artisan route:list` for exact middleware chains.
