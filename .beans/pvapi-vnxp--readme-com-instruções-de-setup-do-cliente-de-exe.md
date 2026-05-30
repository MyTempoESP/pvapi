---
# pvapi-vnxp
title: README com instruções de setup do cliente de exemplo
status: todo
type: task
priority: normal
created_at: 2026-05-30T17:37:53Z
updated_at: 2026-05-30T17:37:58Z
parent: pvapi-outm
---

## Parent

pvapi-outm — Cliente de exemplo Streamlit

## What to build

Criar `Dev/example-client/README.md` documentando como configurar e rodar a aplicação localmente e via Docker. O README deve estar em português brasileiro (mesmo idioma da aplicação) e cobrir:

- Criação de virtualenv (`python -m venv venv && source venv/bin/activate`)
- Instalação de dependências (`pip install -r requirements.txt`)
- Execução local (`streamlit run app.py`)
- Build e execução Docker (`docker build -t pvauto-client . && docker run -p 8501:8501 pvauto-client`)
- Descrição de alto nível das três abas (Checkout, Banco & Saque, Requisição HTTP)

## Acceptance criteria

- [ ] `Dev/example-client/README.md` existe
- [ ] Instruções de virtualenv, pip install e streamlit run corretas e testáveis
- [ ] Instruções de Docker build e run presentes
- [ ] Descrição das três abas em linguagem clara
- [ ] Texto integralmente em pt-BR

## Blocked by

None — can start immediately.

Triage: ready-for-agent
