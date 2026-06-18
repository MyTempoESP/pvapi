package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPickup(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Pickup %d", runID)
	shopEmail := fmt.Sprintf("api-pickup-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")
	_ = shop

	t.Run("create_pickup", func(t *testing.T) {
		result := h.createPickup(t, &shopSession)
		if result.ID == 0 {
			t.Fatal("expected non-zero pickup ID")
		}
	})

	t.Run("list_pickups", func(t *testing.T) {
		result := h.createPickup(t, &shopSession)

		rec := h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		if !containsID(pickups, result.ID) {
			t.Fatalf("pickup %d not found in response: %s", result.ID, rec.Body.String())
		}
	})

	t.Run("update_pickup", func(t *testing.T) {
		p := h.createPickup(t, &shopSession)

		putBody := map[string]any{
			"name":     "Ponto Atualizado",
			"datetime": "2025-06-01T10:00:00",
			"address": map[string]any{
				"street":       "Rua Nova",
				"number":       "200",
				"neighborhood": "Centro",
				"city":         "São Paulo",
				"state":        "SP",
				"CEP":          "01001000",
				"complement":   "Sala 5",
			},
		}
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/pontos/%d", p.ID), &shopSession, putBody)
		assertStatus(t, rec, http.StatusOK)

		// Verify via list
		rec = h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		obj := findObjectByStringField(t, pickups, "name", "Ponto Atualizado")
		addr := asMap(t, obj["address"])
		if asString(t, addr["street"]) != "Rua Nova" {
			t.Fatalf("expected street 'Rua Nova', got %q", asString(t, addr["street"]))
		}
	})

	t.Run("update_pickup_partial_name", func(t *testing.T) {
		p := h.createPickup(t, &shopSession)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/pontos/%d", p.ID), &shopSession, map[string]any{
			"name": "Só o Nome Mudou",
		})
		assertStatus(t, rec, http.StatusOK)

		rec = h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		obj := findObjectByStringField(t, pickups, "name", "Só o Nome Mudou")
		if obj == nil {
			t.Fatal("expected pickup with updated name")
		}
	})

	t.Run("update_pickup_unauthorized", func(t *testing.T) {
		p := h.createPickup(t, &shopSession)
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/pontos/%d", p.ID), nil, map[string]any{
			"name": "Hackeado",
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("update_pickup_cross_shop", func(t *testing.T) {
		shop2Name := fmt.Sprintf("Loja Pickup Cross %d", runID)
		shop2Email := fmt.Sprintf("api-pickup-cross-%d@example.com", runID)
		h.createShop(t, &admin, shop2Name, shop2Email, "11988887778")
		shop2Session := h.login(t, shop2Email, "11988887778")

		p := h.createPickup(t, &shopSession)

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/pontos/%d", p.ID), &shop2Session, map[string]any{
			"name": "Tomado",
		})
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("update_pickup_bad_param", func(t *testing.T) {
		// The gate catches invalid params before the controller,
		// returning 403 because HasPickup finds no match with a non-integer ID.
		rec := h.doJSON(t, http.MethodPut, "/pontos/abc", &shopSession, map[string]any{
			"name": "Invalido",
		})
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("update_pickup_not_found", func(t *testing.T) {
		// The gate returns 403 for non-existent pickup IDs as well.
		rec := h.doJSON(t, http.MethodPut, "/pontos/999999", &shopSession, map[string]any{
			"name": "Inexistente",
		})
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("delete_pickup", func(t *testing.T) {
		p := h.createPickup(t, &shopSession)

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/pontos/%d", p.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if asString(t, body["message"]) != "ponto de entrega deletado com sucesso" {
			t.Fatalf("unexpected message: %v", body["message"])
		}

		// Verify it no longer appears in the list
		rec = h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		if containsID(pickups, p.ID) {
			t.Fatalf("pickup %d should have been deleted, but still appears in list", p.ID)
		}
	})

	t.Run("delete_pickup_unauthorized", func(t *testing.T) {
		p := h.createPickup(t, &shopSession)
		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/pontos/%d", p.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("delete_pickup_cross_shop", func(t *testing.T) {
		shop2Name := fmt.Sprintf("Loja Pickup Cross2 %d", runID)
		shop2Email := fmt.Sprintf("api-pickup-cross2-%d@example.com", runID)
		h.createShop(t, &admin, shop2Name, shop2Email, "11988887779")
		shop2Session := h.login(t, shop2Email, "11988887779")

		p := h.createPickup(t, &shopSession)

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/pontos/%d", p.ID), &shop2Session, "", nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("delete_pickup_not_found", func(t *testing.T) {
		// The gate returns 403 for non-existent pickup IDs, consistent with PUT behavior.
		rec := h.doRequest(t, http.MethodDelete, "/pontos/999999", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("delete_pickup_bad_param", func(t *testing.T) {
		// The gate catches invalid params before the controller,
		// returning 403 because HasPickup finds no match with a non-integer ID.
		rec := h.doRequest(t, http.MethodDelete, "/pontos/abc", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("create_pickup_empty_address", func(t *testing.T) {
		pickupName := fmt.Sprintf("ponto-sem-endereco-%d", time.Now().UnixNano())
		rec := h.doJSON(t, http.MethodPost, "/pontos", &shopSession, map[string]any{
			"name":     pickupName,
			"datetime": "2024-03-25T14:05:43",
			"address": map[string]any{
				"street":       "",
				"number":       "",
				"neighborhood": "",
				"city":         "",
				"state":        "",
				"cep":          "",
				"complement":   "",
			},
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify address is omitted from response
		body := decodeJSON(t, rec)
		if _, exists := body["address"]; exists {
			t.Fatal("expected address to be omitted from response when empty")
		}

		// Verify via list — address should be omitted there too
		rec = h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		obj := findObjectByStringField(t, pickups, "name", pickupName)
		if obj == nil {
			t.Fatal("expected pickup with empty address in list")
		}
		if _, exists := obj["address"]; exists {
			t.Fatal("expected address to be omitted from list response when empty")
		}
	})

	t.Run("create_pickup_partial_address", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/pontos", &shopSession, map[string]any{
			"name":     "ponto-parcial",
			"datetime": "2024-03-25T14:05:43",
			"address": map[string]any{
				"street":       "Rua So a Rua",
				"number":       "",
				"neighborhood": "",
				"city":         "",
				"state":        "",
				"cep":          "",
				"complement":   "",
			},
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("update_pickup_remove_address", func(t *testing.T) {
		p := h.createPickup(t, &shopSession)

		// update address to all-empty
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/pontos/%d", p.ID), &shopSession, map[string]any{
			"name": "Ponto Sem Endereco",
			"address": map[string]any{
				"street":       "",
				"number":       "",
				"neighborhood": "",
				"city":         "",
				"state":        "",
				"cep":          "",
				"complement":   "",
			},
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify via list — address omitted
		rec = h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		obj := findObjectByStringField(t, pickups, "name", "Ponto Sem Endereco")
		if obj == nil {
			t.Fatal("expected updated pickup in list")
		}
		if _, exists := obj["address"]; exists {
			t.Fatal("expected address to be omitted after update to empty")
		}
	})

	t.Run("update_pickup_set_address", func(t *testing.T) {
		// First create with empty address
		pickupName := fmt.Sprintf("ponto-add-set-%d", time.Now().UnixNano())
		rec := h.doJSON(t, http.MethodPost, "/pontos", &shopSession, map[string]any{
			"name":     pickupName,
			"datetime": "2024-03-25T14:05:43",
			"address": map[string]any{
				"street":       "",
				"number":       "",
				"neighborhood": "",
				"city":         "",
				"state":        "",
				"cep":          "",
				"complement":   "",
			},
		})
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		id := asInt(t, body["id"])

		// Update with a full address
		rec = h.doJSON(t, http.MethodPut, fmt.Sprintf("/pontos/%d", id), &shopSession, map[string]any{
			"address": map[string]any{
				"street":       "Rua Incluida",
				"number":       "100",
				"neighborhood": "Bairro Novo",
				"city":         "São Paulo",
				"state":        "SP",
				"cep":          "02002020",
				"complement":   "Apto 1",
			},
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify via list — address present
		rec = h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		obj := findObjectByStringField(t, pickups, "name", pickupName)
		if obj == nil {
			t.Fatal("expected pickup in list")
		}
		addr, ok := obj["address"]
		if !ok {
			t.Fatal("expected address to be present after update")
		}
		addrMap := asMap(t, addr)
		if asString(t, addrMap["street"]) != "Rua Incluida" {
			t.Fatalf("expected street 'Rua Incluida', got %q", asString(t, addrMap["street"]))
		}
	})
}
