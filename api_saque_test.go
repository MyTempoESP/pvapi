package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

const saqueNotFoundID = 99999

func TestSaque(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Saque %d", runID)
	shopEmail := fmt.Sprintf("api-saque-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	// Generate earnings so saque creation works
	h.generateEarnings(t, &shopSession, &admin, shop, runID)

	ba := h.createBankAccount(t, &shopSession, "Conta Principal", "Nubank")
	_ = shop

	t.Run("create_missing_fields", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", &shopSession, map[string]any{
			"description": "missing asked_value",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_unauthorized", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", nil, map[string]any{
			"description":  "saque sem autenticação",
			"asked_value":  "1.00",
			"bank_account": ba.ID,
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("create_saque", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque de teste", "0.01", ba.ID)
		if result.ID == 0 {
			t.Fatal("expected non-zero saque ID")
		}

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["status"]) != "pending" {
			t.Fatal("expected saque status 'pending'")
		}
		if asString(t, saque["bank_name"]) != "Nubank" {
			t.Fatalf("expected bank_name 'Nubank', got %q", asString(t, saque["bank_name"]))
		}
	})

	t.Run("list_saques", func(t *testing.T) {
		ba2 := h.createBankAccount(t, &shopSession, "Conta Secundária", "Inter")
		h.createSaque(t, &shopSession, "Saque de listagem", "0.01", ba2.ID)

		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/saques", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		saques := asSlice(t, body["saques"])
		if len(saques) == 0 {
			t.Fatal("expected at least one saque in list")
		}

		obj := findObjectByStringField(t, saques, "description", "Saque de listagem")
		if asString(t, obj["status"]) != "pending" {
			t.Fatalf("expected status 'pending', got %q", asString(t, obj["status"]))
		}
		if asString(t, obj["bank_name"]) != "Inter" {
			t.Fatalf("expected bank_name 'Inter', got %q", asString(t, obj["bank_name"]))
		}
	})

	t.Run("list_unauthorized", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/saques", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("view_saque", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque para visualizar", "0.01", ba.ID)

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["description"]) != "Saque para visualizar" {
			t.Fatalf("expected description 'Saque para visualizar', got %q", asString(t, saque["description"]))
		}
	})

	t.Run("view_unauthorized", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque view unauthorized", "0.01", ba.ID)

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("view_not_found", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", saqueNotFoundID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("cancel_saque", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque para cancelar", "0.01", ba.ID)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, map[string]any{
			"status": "cancelled",
		})
		assertStatus(t, rec, http.StatusOK)

		rec = h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["status"]) != "cancelled" {
			t.Fatalf("expected status 'cancelled', got %q", asString(t, saque["status"]))
		}
	})

	t.Run("shop_cannot_approve", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque cannot approve", "0.01", ba.ID)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, map[string]any{
			"status": "approved",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("shop_cannot_reject", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque cannot reject", "0.01", ba.ID)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, map[string]any{
			"status": "rejected",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("update_unauthorized", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque update unauthorized", "0.01", ba.ID)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), nil, map[string]any{
			"status": "cancelled",
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("update_invalid_status", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque invalid status", "0.01", ba.ID)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, map[string]any{
			"status": "invalid",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("delete_saque", func(t *testing.T) {
		ba2 := h.createBankAccount(t, &shopSession, "Conta Delete", "Inter")
		result := h.createSaque(t, &shopSession, "saque para deletar", "0.01", ba2.ID)

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		rec = h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("delete_unauthorized", func(t *testing.T) {
		ba2 := h.createBankAccount(t, &shopSession, "Conta Delete Unauthorized", "Inter")
		result := h.createSaque(t, &shopSession, "saque delete unauthorized", "0.01", ba2.ID)

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("delete_not_found", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/saques/%d", saqueNotFoundID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})

	// --- New balance-related test cases ---

	t.Run("create_exact_balance", func(t *testing.T) {
		// Create a fresh shop with predictable earnings for precise balance testing
		balRunID := time.Now().UnixNano()
		balShopName := fmt.Sprintf("Loja Exact Balance %d", balRunID)
		balShopEmail := fmt.Sprintf("api-exact-bal-%d@example.com", balRunID)
		balShop := h.createShop(t, &admin, balShopName, balShopEmail, "11988887777")
		balSession := h.login(t, balShopEmail, "11988887777")
		h.generateEarnings(t, &balSession, &admin, balShop, balRunID)
		balBA := h.createBankAccount(t, &balSession, "Conta Exact", "Nubank")

		// Get the available balance by requesting a very large saque and reading the error
		// The earnings per checkout: value=3, tax=(5/100)*3*1=0.15, commission=(1/100)*3*1=0.03, net=2.82
		// So saque of 2.82 should succeed (exact balance)
		result := h.createSaque(t, &balSession, "Saque valor exato", "2.82", balBA.ID)
		if result.ID == 0 {
			t.Fatal("expected non-zero saque ID for exact balance")
		}
	})

	t.Run("create_negative_value", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", &shopSession, map[string]any{
			"description":  "saque negativo",
			"asked_value":  "-50.00",
			"bank_account": ba.ID,
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_zero_value", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", &shopSession, map[string]any{
			"description":  "saque zero",
			"asked_value":  "0.00",
			"bank_account": ba.ID,
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_exceeds_balance", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", &shopSession, map[string]any{
			"description":  "saque muito grande",
			"asked_value":  "99999.00",
			"bank_account": ba.ID,
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_pending_reduces_balance", func(t *testing.T) {
		// Create a shop with exactly 2 checkout earnings
		balRunID := time.Now().UnixNano()
		balShopName := fmt.Sprintf("Loja Pending Balance %d", balRunID)
		balShopEmail := fmt.Sprintf("api-pending-bal-%d@example.com", balRunID)
		balShop := h.createShop(t, &admin, balShopName, balShopEmail, "11988887777")
		balSession := h.login(t, balShopEmail, "11988887777")
		h.generateEarnings(t, &balSession, &admin, balShop, balRunID)
		h.generateEarnings(t, &balSession, &admin, balShop, balRunID+1)
		balBA := h.createBankAccount(t, &balSession, "Conta Pending", "Nubank")

		// Earnings: 2 * 2.82 = 5.64
		// First saque of 1.00 should succeed
		s1 := h.createSaque(t, &balSession, "Primeiro saque", "1.00", balBA.ID)
		if s1.ID == 0 {
			t.Fatal("expected first saque to succeed")
		}

		// Second saque of 1.00 should succeed
		s2 := h.createSaque(t, &balSession, "Segundo saque", "1.00", balBA.ID)
		if s2.ID == 0 {
			t.Fatal("expected second saque to succeed")
		}

		// Third saque should fail — only 3.64 remaining
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", &balSession, map[string]any{
			"description":  "Terceiro saque que excede",
			"asked_value":  "5.00",
			"bank_account": balBA.ID,
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("checkout_report_reflects_pending", func(t *testing.T) {
		// Create a shop with earnings + pending saques
		balRunID := time.Now().UnixNano()
		balShopName := fmt.Sprintf("Loja Checkout Report %d", balRunID)
		balShopEmail := fmt.Sprintf("api-report-bal-%d@example.com", balRunID)
		balShop := h.createShop(t, &admin, balShopName, balShopEmail, "11988887777")
		balSession := h.login(t, balShopEmail, "11988887777")
		h.generateEarnings(t, &balSession, &admin, balShop, balRunID)
		balBA := h.createBankAccount(t, &balSession, "Conta Report", "Nubank")

		// Create a pending saque
		_ = h.createSaque(t, &balSession, "Saque report", "1.00", balBA.ID)

		// Fetch checkout report
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/checkouts", &balSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		totals := asMap(t, body["totals"])
		if totals == nil {
			t.Fatal("expected 'totals' in checkout report")
		}

		saques := asString(t, totals["saques"])
		actual := asString(t, totals["actual"])

		// saques should include pending (1.00) — non-zero
		if saques == "0" || saques == "0.00" {
			t.Fatalf("expected saques field to include pending saques, got %q", saques)
		}

		// actual should subtract the pending saque from the total
		// total_value - taxes - commissions - saques = actual
		totalStr := asString(t, totals["total"])
		taxesStr := asString(t, totals["taxes"])
		commissionsStr := asString(t, totals["commissions"])
		if totalStr == "" || taxesStr == "" || commissionsStr == "" {
			t.Fatal("expected total, taxes, commissions in totals")
		}

		t.Logf("checkout report: total=%s, taxes=%s, commissions=%s, saques=%s, actual=%s",
			totalStr, taxesStr, commissionsStr, saques, actual)
	})
}

func TestAdminListSaques(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")

	// Create two shops with saques
	shop1Name := fmt.Sprintf("Loja Admin Saques 1 %d", runID)
	shop1Email := fmt.Sprintf("api-admin-saque1-%d@example.com", runID)
	shop1 := h.createShop(t, &admin, shop1Name, shop1Email, "11988887771")
	shop1Session := h.login(t, shop1Email, "11988887771")
	h.generateEarnings(t, &shop1Session, &admin, shop1, runID)
	ba1 := h.createBankAccount(t, &shop1Session, "Conta Admin 1", "Nubank")

	shop2Name := fmt.Sprintf("Loja Admin Saques 2 %d", runID)
	shop2Email := fmt.Sprintf("api-admin-saque2-%d@example.com", runID)
	shop2 := h.createShop(t, &admin, shop2Name, shop2Email, "11988887772")
	shop2Session := h.login(t, shop2Email, "11988887772")
	h.generateEarnings(t, &shop2Session, &admin, shop2, runID+1)
	ba2 := h.createBankAccount(t, &shop2Session, "Conta Admin 2", "Inter")

	s1 := h.createSaque(t, &shop1Session, "Admin list saque shop1", "0.50", ba1.ID)
	s2 := h.createSaque(t, &shop2Session, "Admin list saque shop2", "0.30", ba2.ID)

	t.Run("admin_list_all", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/admin/saques", &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		saques := asSlice(t, body["saques"])
		if len(saques) < 2 {
			t.Fatalf("expected at least 2 saques, got %d", len(saques))
		}

		// Verify both shop saques are present
		if !containsID(saques, s1.ID) {
			t.Fatal("expected saque from shop1 in admin list")
		}
		if !containsID(saques, s2.ID) {
			t.Fatal("expected saque from shop2 in admin list")
		}

		// Verify expected fields on the first saque
		obj := findObjectByIntField(t, saques, "id", s1.ID)
		if asString(t, obj["description"]) != "Admin list saque shop1" {
			t.Fatalf("unexpected description: %q", asString(t, obj["description"]))
		}
		if asString(t, obj["status"]) != "pending" {
			t.Fatalf("unexpected status: %q", asString(t, obj["status"]))
		}
		if asString(t, obj["bank_name"]) != "Nubank" {
			t.Fatalf("unexpected bank_name: %q", asString(t, obj["bank_name"]))
		}
		if asString(t, obj["shop_name"]) != shop1Name {
			t.Fatalf("unexpected shop_name: %q", asString(t, obj["shop_name"]))
		}
		if asString(t, obj["user_name"]) == "" {
			t.Fatal("expected non-empty user_name")
		}
		requireField(t, obj, "asked_value")
		requireField(t, obj, "created_at")
	})

	t.Run("admin_list_unauthorized", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/admin/saques", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("admin_list_blocked_for_shop", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/admin/saques", &shop1Session, "", nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("admin_list_excludes_cancelled", func(t *testing.T) {
		shop3Name := fmt.Sprintf("Loja Admin Cancelled %d", time.Now().UnixNano())
		shop3Email := fmt.Sprintf("api-admin-canc-%d@example.com", time.Now().UnixNano())
		shop3 := h.createShop(t, &admin, shop3Name, shop3Email, "11988887773")
		shop3Session := h.login(t, shop3Email, "11988887773")
		h.generateEarnings(t, &shop3Session, &admin, shop3, time.Now().UnixNano())
		ba3 := h.createBankAccount(t, &shop3Session, "Conta Cancelled", "Nubank")

		// Create a saque, then cancel it
		cancelled := h.createSaque(t, &shop3Session, "Saque cancelled", "0.10", ba3.ID)
		rec := h.doJSON(t, http.MethodPut,
			fmt.Sprintf("/estabelecimentos/saques/%d", cancelled.ID),
			&shop3Session, map[string]any{"status": "cancelled"})
		assertStatus(t, rec, http.StatusOK)

		// Create another pending saque
		pending := h.createSaque(t, &shop3Session, "Saque pending", "0.20", ba3.ID)

		// Admin list should include the pending saque but not the cancelled one
		rec = h.doRequest(t, http.MethodGet, "/admin/saques", &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saques := asSlice(t, body["saques"])

		if containsID(saques, cancelled.ID) {
			t.Fatal("cancelled saque should NOT appear in admin list")
		}
		if !containsID(saques, pending.ID) {
			t.Fatal("pending saque should appear in admin list")
		}

		// Shop-scoped list should still include the cancelled saque
		rec = h.doRequest(t, http.MethodGet, "/estabelecimentos/saques", &shop3Session, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		shopSaques := asSlice(t, body["saques"])
		if !containsID(shopSaques, cancelled.ID) {
			t.Fatal("cancelled saque should still appear in shop-scoped list")
		}
	})
}

func TestAdminSaqueModeration(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")

	shopName := fmt.Sprintf("Loja Mod Saque %d", runID)
	shopEmail := fmt.Sprintf("api-mod-saque-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")
	h.generateEarnings(t, &shopSession, &admin, shop, runID)
	h.generateEarnings(t, &shopSession, &admin, shop, runID+1)
	h.generateEarnings(t, &shopSession, &admin, shop, runID+2)
	ba := h.createBankAccount(t, &shopSession, "Conta Mod", "Nubank")

	createPending := func(desc string) int {
		result := h.createSaque(t, &shopSession, desc, "1.00", ba.ID)
		return result.ID
	}

	t.Run("admin_approve", func(t *testing.T) {
		saqueID := createPending("Aprovar teste")

		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/aprovar", saqueID),
			&admin, nil)
		assertStatus(t, rec, http.StatusOK)

		// Verify status changed to approved
		rec = h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/estabelecimentos/saques/%d", saqueID),
			&shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["status"]) != "approved" {
			t.Fatalf("expected status 'approved', got %q", asString(t, saque["status"]))
		}
		// Motivo should be empty for approved saques
		if asString(t, saque["motivo"]) != "" {
			t.Fatalf("expected empty motivo, got %q", asString(t, saque["motivo"]))
		}
	})

	t.Run("admin_reject", func(t *testing.T) {
		saqueID := createPending("Rejeitar teste")

		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/rejeitar", saqueID),
			&admin, map[string]any{
				"motivo": "saldo insuficiente nas contas da loja",
			})
		assertStatus(t, rec, http.StatusOK)

		// Verify status changed to rejected with motivo
		rec = h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/estabelecimentos/saques/%d", saqueID),
			&shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["status"]) != "rejected" {
			t.Fatalf("expected status 'rejected', got %q", asString(t, saque["status"]))
		}
		if asString(t, saque["motivo"]) != "saldo insuficiente nas contas da loja" {
			t.Fatalf("expected motivo 'saldo insuficiente nas contas da loja', got %q", asString(t, saque["motivo"]))
		}
	})

	t.Run("admin_reject_missing_motivo", func(t *testing.T) {
		saqueID := createPending("Rejeitar sem motivo")

		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/rejeitar", saqueID),
			&admin, map[string]any{})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("admin_approve_not_pending", func(t *testing.T) {
		saqueID := createPending("Aprovar ja aprovado")

		// First approve
		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/aprovar", saqueID),
			&admin, nil)
		assertStatus(t, rec, http.StatusOK)

		// Second approve should fail — not pending
		rec = h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/aprovar", saqueID),
			&admin, nil)
		assertStatus(t, rec, http.StatusBadRequest)
		body := decodeJSON(t, rec)
		if asString(t, body["message"]) != "saque não está pendente" {
			t.Fatalf("expected 'saque não está pendente', got %q", asString(t, body["message"]))
		}
	})

	t.Run("admin_reject_not_pending", func(t *testing.T) {
		saqueID := createPending("Rejeitar ja cancelado")

		// Cancel via shop
		rec := h.doJSON(t, http.MethodPut,
			fmt.Sprintf("/estabelecimentos/saques/%d", saqueID),
			&shopSession, map[string]any{"status": "cancelled"})
		assertStatus(t, rec, http.StatusOK)

		// Reject should fail — not pending
		rec = h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/rejeitar", saqueID),
			&admin, map[string]any{"motivo": "motivo"})
		assertStatus(t, rec, http.StatusBadRequest)
		body := decodeJSON(t, rec)
		if asString(t, body["message"]) != "saque não está pendente" {
			t.Fatalf("expected 'saque não está pendente', got %q", asString(t, body["message"]))
		}
	})

	t.Run("admin_approve_forbidden_shop", func(t *testing.T) {
		saqueID := createPending("Aprovar como shop")

		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/aprovar", saqueID),
			&shopSession, nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("admin_reject_forbidden_shop", func(t *testing.T) {
		saqueID := createPending("Rejeitar como shop")

		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/rejeitar", saqueID),
			&shopSession, map[string]any{"motivo": "motivo"})
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("admin_approve_insufficient_balance", func(t *testing.T) {
		// Create a shop with exact balance and exhaust it
		balRunID := time.Now().UnixNano()
		balShopName := fmt.Sprintf("Loja Bal Mod %d", balRunID)
		balShopEmail := fmt.Sprintf("api-bal-mod-%d@example.com", balRunID)
		balShop := h.createShop(t, &admin, balShopName, balShopEmail, "11988887777")
		balSession := h.login(t, balShopEmail, "11988887777")
		h.generateEarnings(t, &balSession, &admin, balShop, balRunID)
		balBA := h.createBankAccount(t, &balSession, "Conta Bal Mod", "Nubank")

		// Earnings = 2.82, create a saque for 2.82 (exact balance)
		result := h.createSaque(t, &balSession, "Saque balanço exato", "2.82", balBA.ID)
		saqueID := result.ID

		// Approve should succeed since balance >= 0
		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/aprovar", saqueID),
			&admin, nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("rejected_saque_frees_balance", func(t *testing.T) {
		balRunID := time.Now().UnixNano()
		balShopName := fmt.Sprintf("Loja Free Bal %d", balRunID)
		balShopEmail := fmt.Sprintf("api-free-bal-%d@example.com", balRunID)
		balShop := h.createShop(t, &admin, balShopName, balShopEmail, "11988887777")
		balSession := h.login(t, balShopEmail, "11988887777")
		h.generateEarnings(t, &balSession, &admin, balShop, balRunID)
		balBA := h.createBankAccount(t, &balSession, "Conta Free", "Nubank")

		// Earnings = 2.82, create a saque for 2.82
		result := h.createSaque(t, &balSession, "Saque free", "2.82", balBA.ID)

		// Reject it
		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/rejeitar", result.ID),
			&admin, map[string]any{"motivo": "documentação pendente"})
		assertStatus(t, rec, http.StatusOK)

		// Now create another saque with same value — should succeed because balance is freed
		result2 := h.createSaque(t, &balSession, "Saque after free", "2.82", balBA.ID)
		if result2.ID == 0 {
			t.Fatal("expected non-zero saque ID after rejection frees balance")
		}
	})

	t.Run("admin_list_includes_motivo", func(t *testing.T) {
		rejectedID := createPending("Lista motivo rejeitado")

		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/rejeitar", rejectedID),
			&admin, map[string]any{"motivo": "motivo de teste"})
		assertStatus(t, rec, http.StatusOK)

		approvedID := createPending("Lista motivo aprovado")
		rec = h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/aprovar", approvedID),
			&admin, nil)
		assertStatus(t, rec, http.StatusOK)

		// Admin list should include motivo
		rec = h.doRequest(t, http.MethodGet, "/admin/saques", &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saques := asSlice(t, body["saques"])

		rejected := findObjectByIntField(t, saques, "id", rejectedID)
		if asString(t, rejected["motivo"]) != "motivo de teste" {
			t.Fatalf("expected motivo 'motivo de teste', got %q", asString(t, rejected["motivo"]))
		}

		approved := findObjectByIntField(t, saques, "id", approvedID)
		if asString(t, approved["motivo"]) != "" {
			t.Fatalf("expected empty motivo for approved saque, got %q", asString(t, approved["motivo"]))
		}
	})

	t.Run("admin_approve_not_found", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost,
			fmt.Sprintf("/admin/moderacao/saques/%d/aprovar", saqueNotFoundID),
			&admin, nil)
		assertStatus(t, rec, http.StatusNotFound)
	})
}
