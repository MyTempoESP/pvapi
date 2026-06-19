# Cliente de Exemplo — pvauto.linhazero.com

Interface de demonstração em **Streamlit** para consumir a API do pvauto.linhazero.com.
Escrito integralmente em português brasileiro.

## Funcionalidades

### 🪙 Checkout via Token
Fluxo de compra via token de remessa:
1. **API Token** — chave de API (Bearer pv_...) obrigatória para o checkout.
2. **Token de remessa** — informa o UUID do token de remessa e carrega os produtos.
3. **Produtos** — exibe dados completos: preço, desconto (`discount`, `discountType`), código, UUID, categoria, cores, tamanhos, flags de personalização, descrição, especificações, imagens.
4. **Retirada** — escolha do ponto de retirada por produto.
5. **Personalização** — pares nome/valor opcionais por produto (ex.: `Name=Fulano`, `Number=10`).
6. **Confirmação** — `POST /checkouts/confirm` com itens (`[{id, quantity, retirada, personalizacao}]`) e dados do cliente (CPF/telefone — ao menos um obrigatório). Exibe QR codes com preço, desconto e retirada.
7. **Cancelamento** — `DELETE /checkouts/{unique_id}` para cancelar pacotes pendentes.

### 🌐 Requisição HTTP genérica
Envia requisições arbitrárias para qualquer endpoint:
- Métodos: GET, POST, PUT, DELETE, PATCH.
- Autenticação opcional (login automático).
- Corpo JSON e query string editáveis.

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

```bash
docker build -t pvauto-client .
docker run -p 8501:8501 pvauto-client
```

Acesse em `http://localhost:8501`.
