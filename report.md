# Análise de Esforço e Planejamento de Sprints – Back-end (PVAPI)

**Equipe:** 2 desenvolvedores  
**Dedicação:** Horas extras (incluindo sábados) – estimativa de ~11h/dia útil  
**Prazo:** 30 a 45 dias corridos  
**Escopo:** Apenas tarefas de back‑end (Go). Front‑end não foi avaliado.

---

Baseado no mapa:

<img width="2311" height="2715" alt="pvauto" src="https://github.com/user-attachments/assets/5565bccd-7037-4534-9d10-e9f0c8f84c68" />


---

## Premissas para estimativa de dias‑homem (dh)

| Classificação | Dias‑homem por tarefa |
|---------------|-----------------------|
| `[$easy]`     | 1 – 2 dh             |
| `[$average]`  | 3 – 5 dh             |
| `[$medium]`   | 6 – 10 dh            |
| `[$hard]`     | 11 – 16 dh           |

**Capacidade do time:**  
- 2 devs × ~11 h/dia = 22 h/dia útil.  
- Cada sprint de 10 dias úteis → ~220 h (≈ 28 dh/dia? Na verdade 22 h/dia × 10 dias = 220 h, convertido a 1 dh = 8 h = 27,5 dh/sprint). Para simplificar usaremos **dias‑homem de 8 h**, ou seja, 2 devs produzem 2,75 dh por dia útil (22/8).  
- 30 dias corridos ≈ 20 dias úteis → 55 dh disponíveis (20 × 2,75).  
- 45 dias corridos ≈ 30 dias úteis → 82 dh disponíveis.

---

## Backlog total (back‑end apenas)

Agrupado por prioridade (Must / Should / Nice). Valores em dh.

### Must (essencial para MVP)

1. [Bug] logar como monitor  
   **Esforço:** 2 dh  
   **Arquivos tocados:** `models/authmodel/authmodel.go` (claims), `controllers/policy/policy.go` (ajuste de rota)  
   **Justificativa:** Ajuste localizado; reprodução do bug já conhecida.

2. [Bug] Criar ocorrência (formato que vem do front)  
   **Esforço:** 1 dh  
   **Arquivos tocados:** `controllers/report/` (inexistente no chat – criar handler)  
   **Justificativa:** Correção pequena de parse no JSON recebido.

3. [$easy] Listar clientes  
   **Esforço:** 1 dh  
   **Arquivos tocados:** `query/client.go` (já tem `ListShopClientsQuery`), `models/clientmodel/model.go` (método `List` pronto)  
   **Justificativa:** Uso de `WithCollect` já implementado; apenas endpoint não existe.

4. [$easy] Unificar cadastro com `FindOrCreate`  
   **Esforço:** 2 dh  
   **Arquivos tocados:** `models/shopmodel/model.go` (`CreateTx`), `models/consigneemodel/model.go` (`Save`), `models/clientmodel/model.go` (`CreateForCart`) – todos já usam `user.Save` internamente.  
   **Justificativa:** Substituir chamadas diretas por `FindOrCreate` do `models/usermodel/model.go`.

5. [$easy] Alterar queries de inserção para upsert  
   **Esforço:** 1 dh  
   **Arquivos tocados:** `query/shop.go`, `query/consignee.go`, `query/client.go` – adicionar `ON CONFLICT`  
   **Justificativa:** Já existe base; só adicionar cláusula.

6. [$easy] Criar tabela `package_sale`  
   **Esforço:** 2 dh  
   **Arquivos tocados:** `query/consignment.go` (nova query `CREATE TABLE`), `models/consignmentmodel/data.go` (novo struct `PackageSale`)  
   **Justificativa:** Script SQL + struct simples.

7. [$easy] Filtragem por data (pickup)  
   **Esforço:** 1 dh  
   **Arquivos tocados:** `query/pickup.go` (adicionar `WHERE pickup_at >= NOW()`)  
   **Justificativa:** Uma cláusula WHERE.

8. [$easy] Listagem simples de pickup (dropdown)  
   **Esforço:** 1 dh  
   **Arquivos tocados:** `models/pickupmodel/model.go` (já tem `List`), `controllers/pickup/` (criar handler)  
   **Justificativa:** Reuso de função existente.

9. [$average] Criar struct `POSRequest{}`  
   **Esforço:** 3 dh  
   **Arquivos tocados:** `models/consignmentmodel/requests.go` (adicionar campo `PickupLocationID`)  
   **Justificativa:** Estrutura de dados nova, sem lógica complexa.

10. [$average] Modificar controller para aceitar POS  
    **Esforço:** 3 dh  
    **Arquivos tocados:** `controllers/cart/cart.go` (tratar novo campo), `controllers/consignment/` (similar)  
    **Justificativa:** Mudança em parse e rota.

11. [$hard] Modificar funções de checkout para usar referência a pickup  
    **Esforço:** 10 dh  
    **Arquivos tocados:** `models/cartmodel/model.go` (`Exec`, `HoldStock`), `query/cart.go` (`CartHoldStock`), `controllers/cart/cart.go` (fluxo), `models/consignmentmodel/model.go` (`AddProductAction.Exec`)  
    **Justificativa:** É a tarefa mais complexa pois mexe em várias camadas e lógica de transação.

### Should (desejável para entrega)

12. [$medium] Frete (API de cálculo) – **versão simplificada**  
    **Esforço:** 5 dh (reduzido)  
    **Arquivos tocados:** `util/util.go` (função de cálculo fixo/tabela), `models/paymentmodel/requests.go` (adicionar campo `freightCost`)  
    **Justificativa:** Para MVP podemos usar regra fixa (ex.: R$15 para qualquer pedido) em vez de API real.

13. [$average] Modal de endereço da remessa  
    **Esforço:** 3 dh  
    **Arquivos tocados:** `controllers/consignment/` (endpoint GET), `models/consignmentmodel/data.go` (endereço já disponível no struct)  
    **Justificativa:** Apenas serializer e expor dados já existentes.

14. [$medium] Estados de entrega (PENDENTE/RETIRADO/DEVOLVIDO)  
    **Esforço:** 8 dh  
    **Arquivos tocados:** `models/consignmentmodel/data.go` (criar enum), `models/consignmentmodel/model.go` (métodos `SetStatus`), `query/consignment.go` (UPDATE), `gate/consignment.go` (verificar permissões)  
    **Justificativa:** Impacta vários arquivos, mas lógica linear.

15. [$average] Listar pagamentos do cliente  
    **Esforço:** 4 dh  
    **Arquivos tocados:** `query/payment.go` (nova query `ListByClient`), `models/paymentmodel/` (criar `PaymentResource`), `controllers/client/` (handler)  
    **Justificativa:** Consulta simples + struct.

16. [$average] Resumo de compras  
    **Esforço:** 4 dh  
    **Arquivos tocados:** `query/cart.go` (histórico por cliente), `models/cartmodel/model.go` (método `FindByClient`), `controllers/cart/` (endpoint)  
    **Justificativa:** Agregar itens de carrinho já finalizados.

### Nice (postergado)

- Financeiro (saques, totais) – ~20 dh  
- QR Code + verificação – ~15 dh  
- Entrega direta (código rastreio) – ~8 dh  
- Filtros avançados, melhorias – ~12 dh  

---

## Plano de sprints (prazo 45 dias, ~82 dh disponíveis)

### Sprint 1 (dias 1‑10) – Base, correções e pickup

| Tarefa | dh | Arquivos |
|--------|----|----------|
| [Bug] login monitor | 2 | `authmodel/authmodel.go`, `policy/policy.go` |
| [Bug] ocorrência | 1 | `controllers/report/` |
| [easy] listar clientes | 1 | `query/client.go`, `models/clientmodel/model.go` |
| [easy] FindOrCreate | 2 | `models/shopmodel/model.go`, `models/consigneemodel/model.go`, `models/clientmodel/model.go` |
| [easy] upsert queries | 1 | `query/shop.go`, `query/consignee.go`, `query/client.go` |
| [easy] package_sale | 2 | `query/consignment.go`, `models/consignmentmodel/data.go` |
| [easy] filtro data pickup | 1 | `query/pickup.go` |
| [easy] listagem pickup | 1 | `models/pickupmodel/model.go`, `controllers/pickup/` |
| [average] POSRequest | 3 | `models/consignmentmodel/requests.go` |
| [average] modificar controller | 3 | `controllers/cart/cart.go`, `controllers/consignment/` |
| **Total** | **17 dh** | |

**Justificativa:** Todas as tarefas são de baixa/média complexidade e independentes. Podem ser paralelizadas.

### Sprint 2 (dias 11‑20) – Checkout, frete simplificado, modal

| Tarefa | dh | Arquivos |
|--------|----|----------|
| [hard] checkout com pickup | 10 | `models/cartmodel/model.go`, `query/cart.go`, `controllers/cart/cart.go`, `models/consignmentmodel/model.go` |
| [medium] frete simplificado | 5 | `util/util.go`, `models/paymentmodel/requests.go` |
| [average] modal endereço | 3 | `controllers/consignment/`, `models/consignmentmodel/data.go` |
| **Total** | **18 dh** | |

**Justificativa:** A parte crítica (checkout) exige que um dev se dedique integralmente, enquanto o segundo implementa frete e modal simultaneamente. Horas extras são esperadas.

### Sprint 3 (dias 21‑30) – Painel do cliente e estados de entrega

| Tarefa | dh | Arquivos |
|--------|----|----------|
| [medium] estados entrega | 8 | `models/consignmentmodel/data.go`, `models/consignmentmodel/model.go`, `query/consignment.go`, `gate/consignment.go` |
| [average] listar pagamentos | 4 | `query/payment.go`, `models/paymentmodel/`, `controllers/client/` |
| [average] resumo compras | 4 | `query/cart.go`, `models/cartmodel/model.go`, `controllers/cart/` |
| **Total** | **16 dh** | |

**Justificativa:** Estados de entrega é a maior tarefa, mas bem isolada. Listar pagamentos e compras podem ser feitos em paralelo.

### Sprint 4 (dias 31‑40) – Sobra para finalizações e Nice (se prazo for 45 dias)

| Tarefa | dh |
|--------|----|
| [medium] entrega direta simplificada (status) | 5 |
| [medium] QR Code (geração) | 8 |
| [medium] rastreio (código) | 8 (opcional) |
| **Total** | **21 dh** |

Caso o prazo seja **30 dias**, esta sprint será cortada e os itens Nice postergados.

---

## Resumo final do esforço por prazo

| Prazo | Dias úteis | Dh disponíveis | Must + Should essenciais (dh) | Cobertura |
|-------|------------|----------------|-------------------------------|-----------|
| 30 dias | 20 | 55 | Sprint1 (17) + Sprint2 (18) + Sprint3 (16) = 51 | ✅ cobre MVP (Sprints 1‑3) |
| 45 dias | 30 | 82 | 51 + Sprint4 (21) = 72 | ✅ cobre MVP + extras |

**Observações importantes:**

- A tarefa `[$hard]` (checkout) é o maior risco. Deve ser iniciada o mais cedo possível.
- O frete foi simplificado para caber no prazo. Caso queiram API real, adicionar ~8 dh extras (passaria para 95 dh, excedendo 45 dias).
- O QR Code e rastreio são inteiramente postergáveis.
- **Front‑end não foi analisado** e consumirá tempo adicional; não está contabilizado aqui.

---

## Conclusão

Com **2 devs dedicados** (horas extras e sábados), é **factível entregar o MVP** (checkout com ponto de retirada, listagens de clientes/pagamentos/pickups, correção de bugs) em **30 a 45 dias**.  
Para o limite inferior (30 dias), cortar a Sprint 4 (QR Code, entrega direta) – eles serão adiados.  
O caminho crítico é a implementação do **checkout com pickup** e **package_sale**; as demais tarefas podem ser distribuídas em paralelo sem dependências fortes.

**Recomendação:**  
- Foco total em Sprints 1‑3.  
- Revisar diariamente o progresso no checkout.  
- Se houver atraso, eliminar frete (usar valor fixo) e modal de endereço (já retornar dados nos detalhes da remessa).  

---

*Análise gerada em 2026‑05‑16.*
````
