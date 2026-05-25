package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Cliente %d", runID)
	shopEmail := fmt.Sprintf("api-client-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")
	_ = shop

	t.Run("list_clients", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/clientes", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if _, ok := body["clients"]; !ok {
			t.Fatalf("expected 'clients' field in response, got %s", rec.Body.String())
		}
	})
}
