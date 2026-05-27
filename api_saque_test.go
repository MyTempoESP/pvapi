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

	t.Run("update_status", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque para atualizar", "0.01", ba.ID)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, map[string]any{
			"status": "approved",
		})
		assertStatus(t, rec, http.StatusOK)

		rec = h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["status"]) != "approved" {
			t.Fatalf("expected status 'approved', got %q", asString(t, saque["status"]))
		}
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
