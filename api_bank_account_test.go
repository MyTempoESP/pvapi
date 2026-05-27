package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestBankAccount(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja BA %d", runID)
	shopEmail := fmt.Sprintf("api-ba-shop-%d@example.com", runID)
	h.createShop(t, &admin, shopName, shopEmail, "11988887771")
	shopSession := h.login(t, shopEmail, "11988887771")

	shop2Name := fmt.Sprintf("Loja BA2 %d", runID)
	shop2Email := fmt.Sprintf("api-ba-shop2-%d@example.com", runID)
	h.createShop(t, &admin, shop2Name, shop2Email, "11988887772")
	shop2Session := h.login(t, shop2Email, "11988887772")

	t.Run("create_missing_name", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/contasbancarias", &shopSession, map[string]any{
			"bank_name": "Nubank",
			"pix_key":   "12345678900",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_missing_bank_name", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/contasbancarias", &shopSession, map[string]any{
			"name":    "Minha Conta",
			"pix_key": "12345678900",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_unauthorized", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/contasbancarias", nil, map[string]any{
			"name":      "Conta Anônima",
			"bank_name": "Nubank",
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("create", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/contasbancarias", &shopSession, map[string]any{
			"name":      "Conta Principal",
			"bank_name": "Nubank",
			"pix_key":   "12345678900",
			"pix_titular": "Loja Teste",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		conta := asMap(t, body["conta"])
		if asInt(t, conta["id"]) == 0 {
			t.Fatal("expected non-zero bank account ID")
		}
	})

	t.Run("create_no_pix", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/contasbancarias", &shopSession, map[string]any{
			"name":      "Conta TED",
			"bank_name": "Itaú",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		conta := asMap(t, body["conta"])
		if asInt(t, conta["id"]) == 0 {
			t.Fatal("expected non-zero bank account ID")
		}
	})

	t.Run("list", func(t *testing.T) {
		h.createBankAccount(t, &shopSession, "Listagem Conta", "Bradesco")

		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/contasbancarias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		contas := asSlice(t, body["contas"])
		if len(contas) == 0 {
			t.Fatal("expected at least one bank account in list")
		}

		obj := findObjectByStringField(t, contas, "name", "Listagem Conta")
		if asString(t, obj["bank_name"]) != "Bradesco" {
			t.Fatalf("expected bank_name 'Bradesco', got %q", asString(t, obj["bank_name"]))
		}
		if asInt(t, obj["shop_id"]) == 0 {
			t.Fatal("expected shop_id to be present")
		}
	})

	t.Run("list_unauthorized", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/contasbancarias", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("list_isolated_between_shops", func(t *testing.T) {
		h.createBankAccount(t, &shopSession, "Conta Isolada", "Nubank")
		h.createBankAccount(t, &shop2Session, "Conta Outra Loja", "Inter")

		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/contasbancarias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		contas := asSlice(t, body["contas"])

		for _, c := range contas {
			obj := asMap(t, c)
			if asString(t, obj["name"]) == "Conta Outra Loja" {
				t.Fatal("shop should not see another shop's bank account")
			}
		}
	})

	t.Run("update", func(t *testing.T) {
		ba := h.createBankAccount(t, &shopSession, "Conta Atualizar", "Nubank")

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/contasbancarias/%d", ba.ID), &shopSession, map[string]any{
			"name":    "Conta Atualizada",
			"bank_name": "Inter",
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify via list
		rec = h.doRequest(t, http.MethodGet, "/estabelecimentos/contasbancarias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		contas := asSlice(t, body["contas"])
		obj := findObjectByStringField(t, contas, "name", "Conta Atualizada")
		if asString(t, obj["bank_name"]) != "Inter" {
			t.Fatalf("expected bank_name 'Inter', got %q", asString(t, obj["bank_name"]))
		}
	})

	t.Run("update_not_owned", func(t *testing.T) {
		ba := h.createBankAccount(t, &shopSession, "Conta Exclusiva", "Nubank")

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/contasbancarias/%d", ba.ID), &shop2Session, map[string]any{
			"name":    "Conta Hackeada",
			"bank_name": "Inter",
		})
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("update_unauthorized", func(t *testing.T) {
		ba := h.createBankAccount(t, &shopSession, "Conta Update Unauthorized", "Nubank")

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/contasbancarias/%d", ba.ID), nil, map[string]any{
			"name":    "Conta Invadida",
			"bank_name": "Inter",
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("update_not_found", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPut, "/estabelecimentos/contasbancarias/99999", &shopSession, map[string]any{
			"name":    "Inexistente",
			"bank_name": "Nubank",
		})
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("delete", func(t *testing.T) {
		ba := h.createBankAccount(t, &shopSession, "Conta Deletar", "Nubank")

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/contasbancarias/%d", ba.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		// Verify deleted
		rec = h.doRequest(t, http.MethodGet, "/estabelecimentos/contasbancarias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		contas := asSlice(t, body["contas"])
		for _, c := range contas {
			obj := asMap(t, c)
			if asString(t, obj["name"]) == "Conta Deletar" {
				t.Fatal("expected deleted bank account to not appear in list")
			}
		}
	})

	t.Run("delete_not_owned", func(t *testing.T) {
		ba := h.createBankAccount(t, &shopSession, "Conta Exclusiva Del", "Nubank")

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/contasbancarias/%d", ba.ID), &shop2Session, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("delete_unauthorized", func(t *testing.T) {
		ba := h.createBankAccount(t, &shopSession, "Conta Del Unauthorized", "Nubank")

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/contasbancarias/%d", ba.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("delete_not_found", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodDelete, "/estabelecimentos/contasbancarias/99999", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})
}
