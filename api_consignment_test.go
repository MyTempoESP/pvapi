package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestConsignment(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Remessa %d", runID)
	shopEmail := fmt.Sprintf("api-con-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	monitorName := fmt.Sprintf("Monitor R %d", runID)
	monitorEmail := fmt.Sprintf("api-con-monitor-%d@example.com", runID)
	monitor := h.createMonitor(t, &shopSession, monitorName, monitorEmail, "27998209673")

	pickup := h.createPickup(t, &shopSession)

	cat := h.createCategory(t, &shopSession, "lixo-eletronico")
	productName := fmt.Sprintf("Produto R %d", runID)
	prd := h.createProduct(t, &shopSession, productName, cat.ID)
	h.approveProduct(t, &admin, prd.ID)

	cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)

	t.Run("create_consignment", func(t *testing.T) {
		if cons.ID == 0 {
			t.Fatal("expected non-zero consignment ID")
		}
	})

	t.Run("list_consignments", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/remessas", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("consignment %d not found in response: %s", cons.ID, rec.Body.String())
		}
	})

	t.Run("view_consignment", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		c := asMap(t, body["data"])
		if asInt(t, c["id"]) != cons.ID {
			t.Fatalf("expected consignment id %d, got body %s", cons.ID, rec.Body.String())
		}
	})

	t.Run("update_consignment", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"destinationId": pickup.ID,
			"monitorId":     monitor.ID,
			"status":        "CONFERIR - DEVOLUCAO",
			"monitorProfit": 5,
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("add_product_to_consignment", func(t *testing.T) {
		h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, productName)
	})

	_ = shop

	t.Run("list_consignment_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		if !containsID(products, prd.ID) {
			t.Fatalf("product %d not found in consignment response: %s", prd.ID, rec.Body.String())
		}
	})

	t.Run("reject_excessive_quantity", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": productName,
					"quantity":    999,
					"price":       3,
					"color":       "#000000",
					"description": "produto atualizado",
					"category":    "porta-medalhas",
					"discount":    10,
					"featured":    true,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatalf("expected non-200 status for absurd quantity, got body %s", rec.Body.String())
		}
	})

	t.Run("reject_duplicate_product_in_consignment", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": productName,
					"quantity":    1,
					"price":       5,
					"color":       "#000000",
					"description": "produto duplicado",
					"category":    "porta-medalhas",
					"discount":    0,
					"featured":    false,
				},
			},
		})
		assertStatus(t, rec, http.StatusConflict)

		body := decodeJSON(t, rec)
		if code := asString(t, body["code"]); code != "duplicated_resource" {
			t.Fatalf("expected code 'duplicated_resource', got %q", code)
		}
	})

	t.Run("product_stock_reflects_package_quantities", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/produtos", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["produtos"])
		if len(products) == 0 {
			t.Fatal("expected at least one product in listing")
		}

		found := false
		for _, p := range products {
			pm := asMap(t, p)
			if asInt(t, pm["id"]) != prd.ID {
				continue
			}
			found = true
			stock := asInt(t, pm["stock"])
			if stock != 2 {
				t.Fatalf("expected product stock 2 (from package quantity), got %d", stock)
			}
		}
		if !found {
			t.Fatalf("product %d not found in listing", prd.ID)
		}
	})

	t.Run("public_list_endpoint", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/remessas-publicas", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("consignment %d not found in public listing: %s", cons.ID, rec.Body.String())
		}
	})

	t.Run("public_list_by_shop_endpoint", func(t *testing.T) {
		// Create a second shop with its own consignment to verify isolation.
		runID := time.Now().UnixNano()
		otherShopName := fmt.Sprintf("Outra Loja %d", runID)
		otherShopEmail := fmt.Sprintf("api-listby-shop-%d@example.com", runID)
		otherShop := h.createShop(t, &admin, otherShopName, otherShopEmail, "11988887777")
		otherSession := h.login(t, otherShopEmail, "11988887777")

		otherMonitor := h.createMonitor(t, &otherSession,
			fmt.Sprintf("Monitor OB %d", runID),
			fmt.Sprintf("api-listby-monitor-%d@example.com", runID),
			"27998209673")

		otherPickup := h.createPickup(t, &otherSession)
		otherCons := h.createConsignment(t, &otherSession, otherPickup.ID, otherMonitor.ID)

		// Hit the per-shop endpoint for the original shop.
		path := fmt.Sprintf("/estabelecimentos/%s/remessas", shop.Slug)
		rec := h.doRequest(t, http.MethodGet, path, nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])

		// Must contain the original shop's consignment.
		if !containsID(consignments, cons.ID) {
			t.Fatalf("original consignment %d not found in shop listing: %s", cons.ID, rec.Body.String())
		}

		// Must NOT contain the other shop's consignment.
		if containsID(consignments, otherCons.ID) {
			t.Fatalf("other shop's consignment %d should NOT appear in shop listing: %s", otherCons.ID, rec.Body.String())
		}

		// Also verify the other shop's own endpoint returns its consignment.
		otherPath := fmt.Sprintf("/estabelecimentos/%s/remessas", otherShop.Slug)
		otherRec := h.doRequest(t, http.MethodGet, otherPath, nil, "", nil)
		assertStatus(t, otherRec, http.StatusOK)

		otherBody := decodeJSON(t, otherRec)
		otherConsignments := asSlice(t, otherBody["data"])
		if !containsID(otherConsignments, otherCons.ID) {
			t.Fatalf("other shop's consignment %d should appear in its own listing: %s", otherCons.ID, otherRec.Body.String())
		}

		// Must NOT contain the original shop's consignment.
		if containsID(otherConsignments, cons.ID) {
			t.Fatalf("original consignment %d should NOT appear in other shop listing: %s", cons.ID, otherRec.Body.String())
		}
	})
}
