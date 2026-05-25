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
	_ = shop

	t.Run("create_missing_fields", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", &shopSession, map[string]any{
			"description": "missing asked_value and bank_name",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_unauthorized", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", nil, map[string]any{
			"description": "saque sem autenticação",
			"asked_value": "100.00",
			"bank_name":   "Nubank",
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("create_saque", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque de teste", "250.00", "Nubank")
		if result.ID == 0 {
			t.Fatal("expected non-zero saque ID")
		}

		// Verify status via view
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["status"]) != "pending" {
			t.Fatal("expected saque status 'pending'")
		}
	})

	t.Run("list_saques", func(t *testing.T) {
		h.createSaque(t, &shopSession, "Saque de listagem", "250.00", "Nubank")

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
	})

	t.Run("list_unauthorized", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/saques", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("view_saque", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque para visualizar", "300.00", "Nubank")

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		saque := asMap(t, body["saque"])
		if asString(t, saque["description"]) != "Saque para visualizar" {
			t.Fatalf("expected description 'Saque para visualizar', got %q", asString(t, saque["description"]))
		}
	})

	t.Run("view_unauthorized", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque view unauthorized", "300.00", "Nubank")

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("view_not_found", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", saqueNotFoundID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("update_status", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque para atualizar", "250.00", "Nubank")

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
		result := h.createSaque(t, &shopSession, "Saque update unauthorized", "250.00", "Nubank")

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), nil, map[string]any{
			"status": "cancelled",
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("update_invalid_status", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "Saque invalid status", "250.00", "Nubank")

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, map[string]any{
			"status": "invalid",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("delete_saque", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "saque para deletar", "50.00", "Inter")

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		rec = h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("delete_unauthorized", func(t *testing.T) {
		result := h.createSaque(t, &shopSession, "saque delete unauthorized", "50.00", "Inter")

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/saques/%d", result.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("delete_not_found", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/saques/%d", saqueNotFoundID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})
}
