package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCategory(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Categoria %d", runID)
	shopEmail := fmt.Sprintf("api-cat-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")
	_ = shop

	catName := fmt.Sprintf("lixo-eletronico-%d", runID)
	cat := h.createCategory(t, &shopSession, catName)

	t.Run("create_category", func(t *testing.T) {
		if cat.ID == 0 {
			t.Fatal("expected non-zero category ID")
		}
	})

	t.Run("list_categories", func(t *testing.T) {

		rec := h.doRequest(t, http.MethodGet, "/categorias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		categories := asSlice(t, body["data"])
		foundID := findIDByFieldValue(t, categories, "name", catName)
		if foundID != cat.ID {
			t.Fatalf("category id %d not equal to listed id %d", cat.ID, foundID)
		}
	})
}
