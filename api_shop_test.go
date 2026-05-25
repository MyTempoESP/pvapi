package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestShop(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Nova Loja %d", runID)
	updatedName := fmt.Sprintf("Nova Loja Atualizada %d", runID)
	shopEmail := fmt.Sprintf("api-shop-%d@example.com", runID)

	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")

	t.Run("create_shop", func(t *testing.T) {
		if shop.ID == 0 {
			t.Fatal("expected non-zero shop ID")
		}
		if shop.Slug == "" {
			t.Fatal("expected non-empty shop slug")
		}
	})

	t.Run("view_shop", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/%d", shop.ID), &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		s := asMap(t, body["shop"])
		if asInt(t, s["id"]) != shop.ID {
			t.Fatalf("expected shop id %d, got body %s", shop.ID, rec.Body.String())
		}
	})

	t.Run("update_shop", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/%d", shop.ID), &admin, map[string]any{
			"name": updatedName,
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("list_active_shops", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos", &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		shops := asSlice(t, body["shops"])
		if !containsID(shops, shop.ID) {
			t.Fatalf("shop %d not found in response: %s", shop.ID, rec.Body.String())
		}
	})

	t.Run("list_admin_shops", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/all", &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		shops := asSlice(t, body["shops"])
		if len(shops) == 0 {
			t.Fatal("no shops included in admin listing")
		}
	})

	t.Run("deactivate_and_public_list_excludes_inactive", func(t *testing.T) {
		inactiveName := fmt.Sprintf("inactive-%d", runID)
		inactiveEmail := fmt.Sprintf("api-inactive-%d@example.com", runID)

		result := h.createShop(t, &admin, inactiveName, inactiveEmail, "11999999999")

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/estabelecimentos/%d", result.ID), &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)

		// Admin list should still include it
		rec = h.doRequest(t, http.MethodGet, "/estabelecimentos/all", &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		shops := asSlice(t, body["shops"])
		if !containsID(shops, result.ID) {
			t.Fatal("deactivated shop missing from admin listing")
		}

		// Public list must NOT include deactivated shop
		rec = h.doRequest(t, http.MethodGet, "/estabelecimentos", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		shops = asSlice(t, body["shops"])
		if containsID(shops, result.ID) {
			t.Fatal("deactivated shop found in public listing")
		}
	})
}
