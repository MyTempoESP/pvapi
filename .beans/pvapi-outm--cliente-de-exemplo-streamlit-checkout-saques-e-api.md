---
# pvapi-outm
title: Cliente de exemplo Streamlit — checkout, saques e API request
status: done
type: feature
created_at: 2026-05-30T17:31:11Z
updated_at: 2026-05-30T17:31:11Z
---

## Problema

As skills `cart-checkout`, `bank-saque` e `api-request` expõem a API do pvauto.linhazero.com apenas via scripts shell e curl. Não existe uma interface visual para testar ou demonstrar os fluxos de forma integrada. Um operador não-técnico (ou um demonstração para cliente) precisa de uma UI simples que una todos os fluxos num só lugar.

## Solução

Uma aplicação **Streamlit** (`Dev/example-client`) que oferece três fluxos numa única página com navegação por sidebar:

1. **Checkout** — listar produtos, adicionar ao carrinho, escolher pacote, confirmar compra
2. **Banco & Saque** — login, criar conta bancária, solicitar saque
3. **Requisição HTTP genérica** — escape hatch para qualquer endpoint da API

Toda a interface em **português brasileiro**. Aplicação empacotada com virtualenv + Dockerfile.

## Histórias de Usuário

1. Como operador, quero ver a lista de produtos disponíveis, para escolher o que comprar.
2. Como operador, quero adicionar um produto ao carrinho informando ID e quantidade, para montar um pedido.
3. Como operador, quero visualizar os itens do carrinho atual, para conferir antes de finalizar.
4. Como operador, quero ver os pacotes de retirada disponíveis com seus preços e estoque, para escolher o melhor.
5. Como operador, quero selecionar um pacote e confirmar a compra informando valor pago e referência externa, para concluir o checkout.
6. Como operador, quero informar e-mail e nome do cliente durante o checkout, para que o sistema crie o cliente e envie e-mail de boas-vindas.
7. Como operador, quero fazer login com e-mail e senha da loja, para acessar endpoints autenticados.
8. Como operador, quero cadastrar uma conta bancária informando nome, banco, chave PIX e titular, para vincular ao estabelecimento.
9. Como operador, quero solicitar um saque informando descrição, valor e conta bancária, para registrar uma retirada.
10. Como operador, quero enviar requisições HTTP arbitrárias (GET/POST/PUT/DELETE/PATCH) para qualquer endpoint da API, com corpo JSON, query params e autenticação opcional, para testar endpoints não cobertos pelos fluxos específicos.
11. Como operador, quero ver a resposta da API formatada (JSON indentado), para inspecionar o resultado.
12. Como operador, quero alternar entre os três modos (Checkout, Banco & Saque, API Request) via menu lateral, para navegar sem recarregar páginas.
13. Como operador, quero que toda a interface esteja em português brasileiro, para operar sem barreira de idioma.
14. Como devops, quero um Dockerfile que construa e sirva a aplicação, para deploy em qualquer ambiente com Docker.
15. Como dev, quero instalar as dependências via virtualenv e rodar localmente com `streamlit run`, para iterar rapidamente durante desenvolvimento.

## Decisões de Implementação

### Estrutura do projeto

- Único arquivo `app.py` dentro de `Dev/example-client/`. Navegação por `st.sidebar.radio` com três opções: "Checkout", "Banco & Saque", "Requisição HTTP".
- `requirements.txt` com `streamlit` e `requests`.
- `Dockerfile` multi-stage ou single-stage simples expondo porta 8501.
- Virtualenv criado localmente; instruções no README.

### URL base da API

- Constante `BASE_URL = "https://pvauto.linhazero.com"` no topo do `app.py`, facilmente alterável.

### Checkout

- **Seção 1 — Produtos**: botão "Carregar produtos" que faz GET `/produtos` e exibe tabela (ID, nome, preço). Estado em `st.session_state.products`.
- **Seção 2 — Carrinho**: inputs de product_id e quantity lado a lado, botão "Adicionar ao carrinho". ID de sessão do carrinho gerado automaticamente (`cart_<timestamp>`) e salvo em `st.session_state`. Abaixo, botão "Ver carrinho" que faz GET `/carrinho/dados` e exibe itens.
- **Seção 3 — Pacotes**: botão "Carregar pacotes" que faz GET `/carrinho/pacotes` e exibe tabela com ID, nome, preço e estoque. Radio para selecionar pacote.
- **Seção 4 — Confirmar**: inputs para valor_pago, referência externa, e-mail do cliente (opcional), nome do cliente (opcional). Botão "Confirmar compra" que faz POST `/checkouts/confirm`.
- Layout: seções empilhadas verticalmente. Seções 2-4 visíveis desde o início, mostrando estado vazio/mensagem placeholder quando não há dados.

### Banco & Saque

- **Seção 1 — Login**: inputs de e-mail e senha. Botão "Entrar". Token JWT salvo em `st.session_state.auth_token`. Exibe feedback de sucesso/erro.
- **Seção 2 — Conta bancária**: formulário com nome, banco, chave PIX (opcional), titular PIX (opcional). Botão "Criar conta bancária". Após criar, exibe ID da conta criada.
- **Seção 3 — Saque**: input para ID da conta bancária (preenchido automaticamente se criada na seção 2), descrição, valor solicitado. Botão "Solicitar saque".
- Layout: seções empilhadas. Seção 1 sempre visível. Seções 2 e 3 sempre visíveis (podem ser usadas independentemente).

### Requisição HTTP

- Select para método (GET/POST/PUT/DELETE/PATCH).
- Input de endpoint (ex: `/produtos`).
- Checkbox "Autenticar" — se marcado, mostra campos de e-mail e senha e faz login automático antes da requisição.
- Text area para corpo JSON (visível apenas para métodos com body: POST/PUT/PATCH).
- Input para query string.
- Botão "Enviar".
- Resposta exibida em `st.code` com linguagem JSON, incluindo status code.

### Estado da sessão

Chaves no `st.session_state`:
- `cart_session_id`: string, ID da sessão do carrinho
- `auth_token`: string, token JWT
- `products`: lista de dicionários
- `cart_items`: lista de dicionários
- `packages`: lista de dicionários
- `selected_package_id`: int ou None
- `bank_account_id`: int ou None (preenchido após criar conta)

### Português brasileiro

Todos os labels, mensagens, placeholders, textos de botão e feedback em pt-BR. Nomes de variáveis e chaves de código mantidos em inglês (convenção de programação). Respostas da API exibidas como JSON cru (a API retorna em pt-BR ou inglês conforme o servidor; não traduzimos as respostas).

### Docker

- Imagem base: `python:3.12-slim`.
- Copia `app.py` e `requirements.txt`.
- Instala dependências com pip (sem virtualenv dentro do container — desperdício).
- `EXPOSE 8501`.
- `CMD ["streamlit", "run", "app.py", "--server.address=0.0.0.0"]`.

## Decisões de Teste

Não serão escritos testes automatizados para esta aplicação. O escopo é um cliente de exemplo/demonstração. Validação será manual contra a API de produção.

## Fora de Escopo

- Autenticação persistente (refresh token, logout).
- Carrinho persistente entre recarregamentos de página além do session_state do Streamlit.
- Upload de arquivos ou imagens.
- Histórico de requisições.
- Suporte a múltiplos ambientes (staging/produção) via UI — troca-se a constante no código.
- Internacionalização (i18n) — toda a UI é exclusivamente pt-BR.
- Cobertura de testes automatizados.
- Tratamento de erros avançado — exibe o JSON de erro da API e o status code.

## Notas Adicionais

- A API base é `https://pvauto.linhazero.com`. Todos os endpoints documentados nas skills (`cart-checkout`, `bank-saque`, `api-request`) são suportados.
- O carrinho usa session_id arbitrário (não requer autenticação).
- Banco e saques exigem autenticação (JWT via `/login`).
- A aplicação é stateless além do `st.session_state` — não há banco de dados local.

Triage: done
closed_at: 2026-05-30T19:00:00Z

## Resultados

Implementado conforme a especificação:

- **app.py** em `Dev/example-client/app.py` — 3 abas no sidebar (Checkout, Banco & Saque, Requisição HTTP)
- **requirements.txt** — streamlit + requests
- **Dockerfile** — python:3.12-slim, porta 8501
