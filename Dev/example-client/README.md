# Cliente de Exemplo — pvauto.linhazero.com

Interface de demonstração em **Streamlit** para consumir a API do pvauto.linhazero.com.
Escrito integralmente em português brasileiro.

## Funcionalidades

O cliente possui seis abas, acessíveis pelo menu lateral:

### 🛍️ Checkout
Fluxo de compra via listagem de produtos por loja:
1. **Lojas** — lista os estabelecimentos ativos via `GET /estabelecimentos` (opcional, para filtrar produtos).
2. **Produtos disponíveis** — lista os produtos via `GET /produtos`, com filtro por loja. Cada produto exibe seus pacotes de retirada (preço, estoque, ponto de retirada, endereço).
3. **Selecionar pacotes** — escolhe pacotes e quantidades diretamente, sem carrinho intermediário.
4. **Confirmar compra** — finaliza a compra via `POST /checkouts/confirm` com itens (`[{package_id, quantity}]`) e dados do cliente (telefone, CPF — ao menos um obrigatório). A resposta inclui a lista `qr_codes[]` com um QR code por pacote reservado (`{package_id, qr_code}`), renderizados como imagens.

### 🪙 Checkout (Token)
Fluxo de compra via token de remessa:
1. **Token de remessa** — informa o UUID do token de remessa.
2. **Carregar** — `GET /produtos?token={token_de_remessa}` retorna `produtos` (onde `id` é o `package_id`) e `retiradas` (pontos de retirada).
3. **Selecionar produtos** — escolhe quantidades; o `id` de cada produto já é o `package_id`.
4. **Confirmar compra** — `POST /checkouts/confirm` com os `id`s selecionados como `package_id`.

### 🏦 Banco & Saque
Gestão de conta bancária e solicitação de saques (autenticado como estabelecimento):
1. **Login** — autentica-se com e-mail e senha para obter um token JWT.
2. **Conta bancária** — cria e lista contas bancárias associadas ao estabelecimento.
3. **Saque** — solicita e lista saques vinculados a uma conta bancária.

### 🌐 Requisição HTTP genérica
Envia requisições arbitrárias para qualquer endpoint da API:
- Métodos suportados: GET, POST, PUT, DELETE, PATCH.
- Autenticação opcional (login automático antes da requisição).
- Corpo JSON editável para métodos com payload.
- Query string livre para parâmetros de URL.

## Requisitos

- Python 3.12 ou superior
- pip
- (Opcional) Docker

## Setup local

### 1. Criar e ativar um ambiente virtual

```bash
python -m venv venv
source venv/bin/activate
```

No Windows (PowerShell):
```powershell
python -m venv venv
.\venv\Scripts\Activate.ps1
```

### 2. Instalar dependências

```bash
pip install -r requirements.txt
```

### 3. Executar localmente

```bash
streamlit run app.py
```

A aplicação será iniciada em `http://localhost:8501`.

## Execução com Docker

### Build da imagem

```bash
docker build -t pvauto-client .
```

### Executar o container

```bash
docker run -p 8501:8501 pvauto-client
```

Acesse em `http://localhost:8501`.

### Parar o container

```bash
# Descubra o ID ou nome do container
docker ps

# Pare o container
docker stop <container-id>
```

## Observações

- A aplicação se conecta ao servidor de produção (`https://pvauto.linhazero.com`).
- Os tokens de autenticação são armazenados apenas na sessão do navegador (não persistem entre recarregamentos da página).
