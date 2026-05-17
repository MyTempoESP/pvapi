# Análise de Esforço e Planejamento de Sprints – Back-end (PVAPI)

**Equipe:** 2 desenvolvedores (Rodrigo – 11 h/dia, Kerlon – 5 h/dia)  
**Dedicação:** Horas extras, incluindo sábados  
**Prazo:** 30 dias corridos (≈ 20 dias úteis)  
**Capacidade total disponível:** 16 h/dia × 20 dias = 320 h  
**Escopo:** Apenas back‑end (Go). Front‑end não avaliado.

---

Baseado no mapa:

<img width="2311" height="2715" alt="pvauto" src="https://github.com/user-attachments/assets/5565bccd-7037-4534-9d10-e9f0c8f84c68" />


---

## Metodologia

As estimativas foram convertidas de dias‑homem (dh) para **horas‑homem (h)**, com base na calibração fornecida pelo desenvolvedor Rodrigo:

| Classificação | Faixa de horas por tarefa |
|---------------|----------------------------|
| `[$easy]`     | 2‑4 horas                  |
| `[$average]`  | 4‑6 horas                  |
| `[$medium]`   | 6‑10 horas                 |
| `[$hard]`     | 10‑20 horas                |

Para efeito de planejamento, adotamos o valor médio de cada faixa, com acréscimo de **30% de folga (respiro)** para imprevistos, comunicação interna, code review e alinhamento.  
A folga é aplicada no total de cada sprint, não por tarefa.

---

## Capacidade do time (úteis)

| Desenvolvedor | h/dia | Dias úteis (30 dias) | Total h |
|--------------|-------|----------------------|---------|
| Rodrigo      | 11    | 20                   | 220     |
| Kerlon       | 5     | 20                   | 100     |
| **Total**    | 16    | 20                   | 320     |

---

## Backlog reestimado (horas)

| ID | Tarefa | Categoria | h (base) | h c/ folga 30% |
|----|--------|-----------|----------|----------------|
| 1  | [Bug] logar como monitor | $easy | 3 | 4 |
| 2  | [Bug] criar ocorrência (formato front) | $easy | 2 | 3 |
| 3  | [$easy] Listar clientes | $easy | 3 | 4 |
| 4  | [$easy] Unificar cadastro com FindOrCreate | $easy | 4 | 5 |
| 5  | [$easy] Upsert queries inserção | $easy | 3 | 4 |
| 6  | [$easy] Criar tabela package_sale | $easy | 3 | 4 |
| 7  | [$easy] Filtragem por data (pickup) | $easy | 2 | 3 |
| 8  | [$easy] Listagem pickup (dropdown) | $easy | 2 | 3 |
| 9  | [$average] Criar POSRequest | $average | 5 | 7 |
| 10 | [$average] Modificar controller cart/consignment para POS | $average | 5 | 7 |
| 11 | [$hard] Checkout com referência a pickup | $hard | 16 | 21 |
| 12 | [$medium] Frete simplificado | $medium | 8 | 10 |
| 13 | [$average] Modal endereço da remessa | $average | 5 | 7 |
| 14 | [$medium] Estados de entrega (PENDENTE/RETIRADO/DEVOLVIDO) | $medium | 10 | 13 |
| 15 | [$average] Listar pagamentos do cliente | $average | 5 | 7 |
| 16 | [$average] Resumo de compras | $average | 5 | 7 |
| 17 | [$hard] Geração e verificação de código QR | $hard | 16 | 21 |
|     | **Total** | | **~97 h** | **~129 h** |

*Nota: Folga de 30% inclui lidar com dúvidas, refatorações pequenas e comunicação em paralelo.*

---

## Alocação por sprint (cada sprint = 10 dias úteis → ~160 h disponíveis)

### Sprint 1 (dias 1‑10) – Base, correções, pickups

| Tarefa | h c/ folga |
|--------|------------|
| 1. Bug login monitor | 4 |
| 2. Bug ocorrência | 3 |
| 3. Listar clientes | 4 |
| 4. FindOrCreate | 5 |
| 5. Upsert queries | 4 |
| 6. Criar tabela package_sale | 4 |
| 7. Filtro data pickup | 3 |
| 8. Listagem pickup | 3 |
| 9. POSRequest | 7 |
| 10. Modificar controller | 7 |
| **Subtotal** | **44 h** |
| Respiro alocado (não utilizado) | 116 h |
| **Total disponível** | **160 h** |

*Sprint folgada: sobra tempo para alinhamento com front‑end e eventuais dificuldades.*

### Sprint 2 (dias 11‑20) – Checkout, frete, modal

| Tarefa | h c/ folga |
|--------|------------|
| 11. Checkout c/ pickup (hard) | 21 |
| 12. Frete simplificado | 10 |
| 13. Modal endereço | 7 |
| **Subtotal** | **38 h** |
| **Total disponível** | **160 h** → respiro extra de **122 h** para testes e ajustes. |

*Obs: O checkout é o maior risco. Caso o dev Rodrigo se dedique integralmente (11 h/dia) por 2 dias (22h), o restante da sprint fica para Kerlon (5 h/dia) concluir frete e modal.*

### Sprint 3 (dias 21‑30) – Estados, pagamentos, resumo + Financeiro

| Tarefa | h c/ folga |
|--------|------------|
| 14. Estados de entrega | 13 |
| 15. Listar pagamentos do cliente | 7 |
| 16. Resumo de compras | 7 |
| 17. Geração e verificação de QR Code | 21 |
| 18. Modelo de saque | 10 |
| 19. Solicitar saque | 10 |
| 20. Listar saques | 7 |
| 21. Painel financeiro | 8 |
| 22. Aprovar saque | 4 |
| **Subtotal** | **87 h** |
| **Total disponível** | **160 h** → **73 h** de respiro. |

---

## Resumo do esforço vs. capacidade

| Período | Hh necessárias (c/ folga) | Hh disponíveis (320) | Saldo (respiro) |
|---------|---------------------------|----------------------|-----------------|
| Sprint 1 | 44 | 160 | 116 |
| Sprint 2 | 38 | 160 | 122 |
| Sprint 3 | 87 | 160 | 73 |
| **Total** | **169 h** | **320 h** | **151 h** |

**Conclusão:**  
O MVP cabe confortavelmente nas 3 sprints (30 dias úteis), com grande margem para respiro, reuniões, suporte ao front‑end e imprevistos.

---

## Tarefas financeiras sugeridas

| ID | Tarefa | Categoria | h base | h c/ 30% folga |
|----|--------|-----------|--------|----------------|
| 18 | [$medium] Criar modelo de saque (tabela, queries, modelo) | medium | 8 | 10 |
| 19 | [$medium] Endpoint solicitar saque (criação, validação de saldo) | medium | 8 | 10 |
| 20 | [$average] Listar saques do usuário (histórico) | average | 5 | 7 |
| 21 | [$average] Painel financeiro: totais (receita líquida, comissões, saldo disponível) | average | 6 | 8 |
| 22 | [$easy] Endpoint de aprovação/rejeição de saque (admin) | easy | 3 | 4 |
|     | **Total** | | **30 h** | **39 h** |

---

## Dependências críticas

- **Checkout (tarefa 11)** deve ser iniciado no início da Sprint 2, pois impacta diretamente o fluxo de pagamento e criação de remessa.  
- **POSRequest (tarefa 9)** deve estar pronto antes do checkout (pode ser concluído na Sprint 1).  
- **Estados de entrega (tarefa 14)** pode ser iniciado em paralelo com checkout, mas não impacta o MVP se entregue até o final da Sprint 3.  
- **Painel financeiro** precisa consultar totais das vendas concluídas e dos saques realizados – pode ser desenvolvido em paralelo com os endpoints de saque.

---

## Recomendações pós‑MVP (Nice – postergado)

- Entrega direta (código rastreio) – ~12 h  
- Filtros avançados – ~18 h  

Total ~30 h. Com a capacidade disponível após MVP (151 h de respiro), parte dessas tarefas pode ser absorvida **dentro do mesmo prazo de 30 dias**, caso o time opte.

---

**Planejamento gerado em 2026‑05‑16 – versão refatorada com horas.**
