# Cliente de Exemplo — pvauto.linhazero.com

Interface de demonstração em **Streamlit** para consumir a API do pvauto.linhazero.com.
Escrito integralmente em português brasileiro.

## Funcionalidades

O cliente possui três abas, acessíveis pelo menu lateral:

### 🛍️ Checkout
Fluxo completo de compra:
1. **Lojas** — lista os estabelecimentos ativos via `GET /estabelecimentos` (opcional, para filtrar produtos).
2. **Produtos disponíveis** — lista os produtos cadastrados via `GET /produtos`, com filtro por loja. Cada produto exibe seus pacotes de retirada (preço, estoque, ponto de retirada, endereço).
3. **Selecionar pacotes** — escolhe pacotes e quantidades diretamente, sem carrinho intermediário.
4. **Confirmar compra** — finaliza a compra via `POST /checkouts/confirm` com valor pago, referência externa, itens (`[{package_id, quantity}]`) e dados opcionais do cliente (e-mail e nome). Ao confirmar, um e-mail de confirmação é enviado se o e-mail do cliente for informado.

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
