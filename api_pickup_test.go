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
}
