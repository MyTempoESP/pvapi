# PvAuto — API de Integração

**Base URL:** `https://pvauto.linhazero.com`

## Fluxo Principal

```
1. POST /carrinho/adicionar    → cria/reutiliza carrinho com session_id
2. GET  /carrinho/pacotes       → consulta pacotes disponíveis
3. POST /checkouts/confirm      → finaliza checkout com package_ids
```

O `session_id` é uma string opaca fornecida pelo cliente. Carrinhos já finalizados (prefixo `FINISHED-`) são ignorados.

---

## Autenticação

Nenhuma. Rotas públicas.

---

## Endpoints

### Carrinho

#### `POST /carrinho/adicionar`
Adiciona um produto ao carrinho. Cria ou reutiliza um carrinho ativo vinculado ao `session_id`.

**Request body (`application/json`):**

| Campo            | Tipo    | Obrigatório | Descrição                            |
|------------------|---------|-------------|--------------------------------------|
| `product_id`     | integer | sim         | ID do produto                        |
| `quantity`       | integer | sim         | Quantidade (mín. 1, padrão 1)        |
| `personalization`| object  | não         | Dados de personalização (opcional)   |
| `session_id`     | string  | sim         | Identificador de sessão do cliente   |

**Response `200`:**
```json
{ "success": true, "message": "produto adicionado!" }
```

---

#### `GET /carrinho/dados`
Retorna os itens, subtotais e descontos do carrinho.

**Query params:**

| Parâmetro    | Tipo   | Obrigatório | Descrição                   |
|--------------|--------|-------------|-----------------------------|
| `session_id` | string | sim         | Identificador de sessão     |

**Response `200` — `RespostaCarrinho`:**

| Campo                  | Tipo                              | Descrição                             |
|------------------------|-----------------------------------|---------------------------------------|
| `success`              | boolean                           |                                       |
| `message`              | string                            |                                       |
| `items`                | array de `ItemCarrinho`           | Lista de itens no carrinho            |
| `total`                | number                            | Valor total                           |
| `item_count`           | integer                           | Quantidade de itens distintos         |
| `total_with_discount`  | number                            | Valor total com descontos             |
| `total_discount`       | number                            | Valor total dos descontos             |
| `session_id`           | string                            |                                       |

**Cada `ItemCarrinho`:**

| Campo                    | Tipo    | Descrição                              |
|--------------------------|---------|----------------------------------------|
| `id`                     | integer | ID do item no carrinho                 |
| `product_id`             | integer | ID do produto                          |
| `shop_id`                | integer | ID da loja                             |
| `product_name`           | string  | Nome do produto                        |
| `product_image`          | string  | URL da imagem                          |
| `shop_name`              | string  | Nome da loja                           |
| `quantity`               | integer | Quantidade                             |
| `price`                  | number  | Preço unitário                         |
| `discount`               | number  | Desconto percentual                    |
| `subtotal`               | number  | Preço × quantidade                     |
| `price_with_discount`    | number  | Preço unitário com desconto            |
| `subtotal_with_discount` | number  | Subtotal com desconto                  |

**Errors:** `400` (requisição inválida), `404` (carrinho não encontrado).

---

#### `GET /carrinho/pacotes`
Lista opções de pacote (remessa) para os itens do carrinho, agrupadas por produto.

**Query params:**

| Parâmetro    | Tipo   | Obrigatório | Descrição                   |
|--------------|--------|-------------|-----------------------------|
| `session_id` | string | sim         | Identificador de sessão     |

**Response `200` — `RespostaPacotes`:**

| Campo     | Tipo                                   | Descrição                      |
|-----------|----------------------------------------|--------------------------------|
| `success` | boolean                                |                                |
| `message` | string                                 |                                |
| `products`| array de `PacotesPorProduto`           | Pacotes agrupados por produto  |

**Cada `PacotesPorProduto`:**

| Campo          | Tipo                         | Descrição                      |
|----------------|------------------------------|--------------------------------|
| `product_id`   | integer                      | ID do produto                  |
| `product_name` | string                       | Nome do produto                |
| `packages`     | array de `OpcaoPacote`       | Opções de pacote disponíveis   |

**Cada `OpcaoPacote`:**

| Campo            | Tipo              | Descrição                                   |
|------------------|-------------------|---------------------------------------------|
| `package_id`     | integer           | ID do pacote (usar no checkout)             |
| `consignment_id` | integer           | ID da remessa                               |
| `price`          | number            | Preço do pacote                             |
| `discount`       | number            | Desconto                                    |
| `stock`          | integer           | Estoque disponível                          |
| `pickup`         | `PontoRetirada`   | Ponto de retirada                           |

**`PontoRetirada`:**

| Campo     | Tipo      | Descrição                  |
|-----------|-----------|----------------------------|
| `id`      | integer   | ID do ponto                |
| `name`    | string    | Nome do ponto              |
| `date`    | string    | Data/hora da retirada      |
| `address` | `Endereco`| Endereço completo          |

**`Endereco`:** `street`, `number`, `neighborhood`, `city`, `state`, `cep`, `complement`.

> Produtos sem pacotes disponíveis aparecem com `packages: []`.

**Error:** `400`.

---

#### `DELETE /carrinho/remover/{item}`
Remove um item do carrinho.

**Path params:**

| Parâmetro | Tipo    | Obrigatório | Descrição               |
|-----------|---------|-------------|-------------------------|
| `item`    | integer | sim         | ID do item no carrinho  |

**Query params:**

| Parâmetro    | Tipo   | Obrigatório | Descrição                   |
|--------------|--------|-------------|-----------------------------|
| `session_id` | string | sim         | Identificador de sessão     |

**Response `200`:** `{ "success": true, "message": "..." }`

---

#### `PUT /carrinho/atualizar/{item}`
Altera a quantidade de um item no carrinho.

**Path params:**

| Parâmetro | Tipo    | Obrigatório | Descrição               |
|-----------|---------|-------------|-------------------------|
| `item`    | integer | sim         | ID do item no carrinho  |

**Query params:**

| Parâmetro    | Tipo   | Obrigatório | Descrição                   |
|--------------|--------|-------------|-----------------------------|
| `session_id` | string | sim         | Identificador de sessão     |

**Request body:**

| Campo      | Tipo    | Obrigatório | Descrição              |
|------------|---------|-------------|------------------------|
| `quantity` | integer | sim         | Nova quantidade (≥1)   |

**Response `200`:** `{ "success": true, "message": "..." }`

---

### Checkout

#### `POST /checkouts/confirm`
Finaliza o carrinho, reserva estoque e gera QR code para retirada.

**Idempotente:** Mesmo par `(session_id, external_reference)` retorna o mesmo `unique_id` sem duplicar.

**Query params:**

| Parâmetro    | Tipo   | Obrigatório | Descrição                   |
|--------------|--------|-------------|-----------------------------|
| `session_id` | string | sim         | Identificador de sessão     |

**Request body (`application/json`):**

| Campo                | Tipo                         | Obrigatório | Descrição                                       |
|----------------------|------------------------------|-------------|-------------------------------------------------|
| `value_paid`         | string                       | sim         | Valor pago (string decimal, ex: `"600"`)       |
| `external_reference` | string                       | sim         | Referência externa única (chave de idempotência)|
| `items`              | array de `ItemCheckout`      | sim         | Mín. 1 item                                     |

**Cada `ItemCheckout`:**

| Campo        | Tipo    | Obrigatório | Descrição                            |
|--------------|---------|-------------|--------------------------------------|
| `package_id` | integer | sim         | ID do pacote (via GET /carrinho/pacotes) |
| `quantity`   | integer | sim         | Quantidade deste pacote              |

**Response `200` — `RespostaCheckout`:**

| Campo       | Tipo    | Descrição                                |
|-------------|---------|------------------------------------------|
| `success`   | boolean | `true`                                   |
| `message`   | string  | `"checkout confirmed"`                   |
| `unique_id` | string  | UUID da transação                        |
| `status`    | string  | `"paid"`                                 |
| `qr_code`   | string  | QR code em data URI (PNG base64)         |

**Errors:**

| Status | Significado                |
|--------|----------------------------|
| `400`  | Requisição inválida         |
| `404`  | Carrinho não encontrado/ finalizado |
| `409`  | Estoque insuficiente        |
| `422`  | Carrinho vazio / pacote não encontrado |

---

### Remessas

#### `GET /remessas-publicas`
Lista remessas/consignments ativas com status aprovado.

**Response `200` — `RespostaListaRemessas`:**

| Campo     | Tipo                              | Descrição                  |
|-----------|-----------------------------------|----------------------------|
| `success` | boolean                           |                            |
| `message` | string                            | `"remessas listadas com sucesso"` |
| `data`    | array de `RecursoRemessa`         | Lista de remessas          |

**Cada `RecursoRemessa`:**

| Campo               | Tipo    | Descrição                          |
|---------------------|---------|------------------------------------|
| `id`                | integer | ID da remessa                      |
| `commission_type`   | string  | `fixed` ou `percentage`            |
| `commission_value`  | number  | Valor da comissão                  |
| `code`              | string  | Código                             |
| `slug`              | string  | Slug identificador                 |
| `status`            | string  | Status                             |
| `destination`       | string  | Nome do ponto de retirada          |
| `responsibleMonitor`| string  | Nome do monitor responsável        |
| `created_at`        | string  | ISO 8601 (date-time)               |

---

#### `GET /estabelecimentos/{shop}/remessas`
Lista remessas de uma loja específica (por slug).

**Path params:**

| Parâmetro | Tipo   | Obrigatório | Descrição                         |
|-----------|--------|-------------|-----------------------------------|
| `shop`    | string | sim         | Slug da loja (identificador textual) |

**Response `200`:** Mesmo formato de `RespostaListaRemessas`.

**Error:** `404` (loja não encontrada).

---

### Produtos

#### `GET /produtos`
Lista produtos aprovados do catálogo público com filtros opcionais. Parâmetros multivalorados (repetir `?shop=1&shop=2`).

**Query params:**

| Parâmetro      | Tipo            | Obrigatório | Descrição                                    |
|----------------|-----------------|-------------|----------------------------------------------|
| `shop`         | integer[]       | não         | ID(s) de loja (ex: `?shop=1&shop=2`)         |
| `consignment`  | integer[]       | não         | ID(s) de remessa (ex: `?consignment=3`)      |

> Quando filtrado por `consignment`, cada produto inclui o campo `consignments` (array de IDs das remessas que o contêm).

**Response `200` — `RespostaListaProdutos`:**

| Campo     | Tipo                              | Descrição              |
|-----------|-----------------------------------|------------------------|
| `success` | boolean                           |                        |
| `message` | string                            | `"Produtos listados com sucesso!"` |
| `produtos`| array de `RecursoProduto`         | Lista de produtos      |

**Cada `RecursoProduto`:**

| Campo                      | Tipo                              | Descrição                                   |
|----------------------------|-----------------------------------|---------------------------------------------|
| `id`                       | integer                           | ID do produto                               |
| `name`                     | string                            | Nome                                        |
| `stock`                    | integer                           | Estoque                                     |
| `price`                    | number                            | Preço                                       |
| `discount`                 | number                            | Desconto                                    |
| `featured`                 | boolean                           | Em destaque                                 |
| `code`                     | string                            | Código                                      |
| `uuid`                     | string                            | UUID                                        |
| `platform_cut`             | number                            | Repasse da plataforma                       |
| `platform_cut_type`        | string                            | `fixed` ou `percentage`                     |
| `colors`                   | string[]                          | Cores disponíveis                           |
| `sizes`                    | string[]                          | Tamanhos disponíveis                        |
| `allowCustomName`          | boolean                           | Permite nome personalizado                  |
| `allowCustomModality`      | boolean                           | Permite modalidade personalizada            |
| `allowCustomColorSelection`| boolean                           | Permite seleção de cor personalizada        |
| `moderation_status`        | string                            | Status de moderação                         |
| `category`                 | string                            | Categoria                                   |
| `category_id`              | integer                           | ID da categoria                             |
| `imageUrl`                 | string                            | URL da imagem principal                     |
| `description`              | string                            | Descrição                                   |
| `specificationsImages`     | string[]                          | Imagens de especificações                   |
| `descriptionImages`        | string[]                          | Imagens de descrição                        |
| `deliveryImages`           | string[]                          | Imagens de entrega                          |
| `specifications`           | array de `EspecificacaoProduto`   | Especificações (name + value)               |
| `consignments`             | integer[]                         | IDs das remessas (presente apenas com filtro `consignment`) |

**`EspecificacaoProduto`:** `{ "name": string, "value": string }`

---

#### `GET /produtos/{product}`
Detalhes de um produto específico.

**Path params:**

| Parâmetro | Tipo    | Obrigatório | Descrição    |
|-----------|---------|-------------|--------------|
| `product` | integer | sim         | ID do produto|

**Response `200` — `RespostaProduto`:**

| Campo     | Tipo              | Descrição                  |
|-----------|-------------------|----------------------------|
| `success` | boolean           |                            |
| `message` | string            | `"produto encontrado!"`    |
| `data`    | `RecursoProduto`  | Dados do produto           |

**Error:** `404` (produto não encontrado).

---

## Schemas Compartilhados

### Genéricos

**`RespostaOk`:**
```json
{ "success": true, "message": "string" }
```

**`RespostaErro`:**
```json
{ "success": false, "message": "string" }
```

### Endereço

| Campo        | Tipo    |
|--------------|---------|
| `street`     | string  |
| `number`     | string  |
| `neighborhood`| string |
| `city`       | string  |
| `state`      | string  |
| `cep`        | string  |
| `complement` | string  |

---

## Códigos de Erro

| HTTP | Significado            | Observação                        |
|------|------------------------|-----------------------------------|
| 400  | Requisição inválida    | Campos obrigatórios ausentes, etc.|
| 404  | Não encontrado         | Carrinho, loja ou produto         |
| 409  | Conflito               | Estoque insuficiente              |
| 422  | Entidade não processável| Carrinho vazio / pacote inválido  |
