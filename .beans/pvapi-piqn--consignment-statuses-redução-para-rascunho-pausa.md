---
# pvapi-piqn
title: Consignment statuses redução para RASCUNHO, PAUSADO, LIBERADO, ARQUIVADO
status: todo
type: feature
priority: normal
created_at: 2026-06-19T03:52:24Z
updated_at: 2026-06-19T03:52:29Z
---

## Problem Statement

O ciclo de vida das remessas (consignments) atualmente possui 5 statuses: RASCUNHO, EM TRANSITO, RECEBIDO, CONFERIDO, ARQUIVADO. Os statuses EM TRANSITO, RECEBIDO e CONFERIDO representam granularidade desnecessária para o volume atual (~2 usuários). Manter múltiplos statuses ativos adiciona complexidade sem benefício operacional.

## Solution

Reduzir os statuses de remessa para 4: RASCUNHO, PAUSADO, LIBERADO, ARQUIVADO, onde apenas LIBERADO é considerado "ativo" (visível em listagens públicas). PAUSADO serve como ocultação temporária de uma remessa já liberada, permitindo que o lojista faça ajustes sem removê-la definitivamente.

## User Stories

1. Como lojista, quero criar uma remessa em RASCUNHO, para prepará-la antes de publicar.
2. Como lojista, quero liberar uma remessa (RASCUNHO → LIBERADO), para torná-la visível ao público.
3. Como lojista, quero pausar uma remessa (LIBERADO → PAUSADO), para ocultá-la temporariamente enquanto faço ajustes.
4. Como lojista, quero reativar uma remessa pausada (PAUSADO → LIBERADO), para voltar a exibi-la.
5. Como lojista, quero arquivar uma remessa de qualquer status, para removê-la permanentemente.
6. Como lojista, quero poder editar produtos em uma remessa PAUSADA, para fazer os ajustes necessários antes de reativá-la.
7. Como lojista, quero que remessas PAUSADAS não apareçam em minhas listagens ativas, para não confundir com remessas publicadas.
8. Como cliente público, quero ver apenas remessas LIBERADAS nas listagens, para navegar somente no que está disponível.
9. Como cliente com link direto, quero receber uma mensagem clara "remessa temporariamente indisponível" ao acessar uma remessa PAUSADA, para saber que ela pode voltar.
10. Como cliente com link direto, quero receber "remessa arquivada, não está mais disponível" ao acessar uma remessa ARQUIVADA, para saber que ela foi removida definitivamente.

## Implementation Decisions

- **Status constants**: Substituir EM TRANSITO, RECEBIDO, CONFERIDO por PAUSADO, LIBERADO. Manter RASCUNHO e ARQUIVADO.
- **Visibilidade pública (consignments_active view)**: Apenas LIBERADO é ativo. O filtro da view passa a ser `WHERE status NOT IN ('RASCUNHO', 'PAUSADO', 'ARQUIVADO')`.
- **Sem validação de transição**: Qualquer status pode ser alterado para qualquer outro, sem restrições de máquina de estados.
- **Endpoints de listagem**: Todas as queries que usam `consignments_active` serão atualizadas indiretamente via a view. Queries com filtro hardcoded `status != 'ARQUIVADO'` também excluirão PAUSADO.
- **Endpoint público de produtos por código**: ARQUIVADO retorna HTTP 410. PAUSADO retorna HTTP 404 com mensagem "remessa temporariamente indisponível". LIBERADO permite acesso normal.
- **Remoção de produtos**: Bloqueada apenas para ARQUIVADO (comportamento atual mantido). PAUSADO permite remoção para viabilizar ajustes.
- **Migration**: Novo par de migrações para recriar a view `consignments_active`. Sem migração de dados — linhas com status antigos tornam-se naturalmente invisíveis.
- **Nenhum CHECK constraint novo no banco** — consistência com o comportamento atual (status é texto livre validado apenas na camada de aplicação via request DTO, e mesmo ali sem enum rígido para o status da remessa em si).

## Testing Decisions

- Atualizar fixtures de teste que referenciam statuses removidos (EM TRANSITO, RECEBIDO, CONFERIDO) para usar os novos (LIBERADO, PAUSADO).
- Testes de visibilidade: verificar que RASCUNHO, PAUSADO e ARQUIVADO não aparecem em listagens públicas; LIBERADO sim.
- Testes de listagens autenticadas: verificar que todos os statuses são visíveis para shop e consignee.
- Teste de endpoint público: verificar 404 para PAUSADO e 410 para ARQUIVADO.
- Testes existentes que validam comportamento de RASCUNHO e ARQUIVADO devem continuar passando.
- Seguir padrão existente: testes de integração via API (api_consignment_test.go, api_product_test.go).

## Out of Scope

- Migração de dados de linhas existentes com status antigos.
- Validação de transições de status (máquina de estados).
- Alteração no modelo de pacotes/produtos (packages.status "active"/"paused" permanece inalterado).
- Interface de usuário — apenas API.

## Further Notes

- Os statuses EM TRANSITO, RECEBIDO e CONFERIDO serão removidos dos constantes, queries e testes. Linhas no banco com esses valores permanecerão inalteradas e serão invisíveis nas listagens, mas ainda acessíveis por ID via endpoints autenticados.

Triage: ready-for-agent
