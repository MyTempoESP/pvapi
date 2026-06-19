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

	t.Run("consignment_has_name", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		data := asMap(t, body["data"])
		if name := asString(t, data["name"]); name != "Remessa Teste" {
			t.Fatalf("expected name 'Remessa Teste', got %q", name)
		}
	})

	t.Run("update_consignment_name", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"name": "Nome Atualizado",
		})
		assertStatus(t, rec, http.StatusOK)

		// verify
		rec2 := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		body := decodeJSON(t, rec2)
		data := asMap(t, body["data"])
		if name := asString(t, data["name"]); name != "Nome Atualizado" {
			t.Fatalf("expected name 'Nome Atualizado', got %q", name)
		}

		// revert for downstream tests
		h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"name": "Remessa Teste",
		})
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
		code := h.getConsignmentCode(t, &shopSession, cons.ID)

		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/produtos?token=%s", code),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["produtos"])
		if len(products) == 0 {
			t.Fatal("expected at least one product in listing")
		}

		found := false
		for _, p := range products {
			pm := asMap(t, p)
			// id is now package_id, not product_id
			// We match by name instead
			if asString(t, pm["name"]) != productName {
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

		// each consignment must include a shop object
		for _, c := range consignments {
			cm := asMap(t, c)
			shopObj := asMap(t, cm["shop"])
			if asInt(t, shopObj["id"]) == 0 {
				t.Fatalf("consignment %v has no shop.id", cm["id"])
			}
			if asString(t, shopObj["name"]) == "" {
				t.Fatalf("consignment %v has no shop.name", cm["id"])
			}
		}
	})

	t.Run("public_list_filtered_by_shop", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/remessas-publicas?loja=%d", shop.ID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("consignment %d not found in filtered listing: %s", cons.ID, rec.Body.String())
		}

		for _, c := range consignments {
			cm := asMap(t, c)
			shopObj := asMap(t, cm["shop"])
			if asInt(t, shopObj["id"]) != shop.ID {
				t.Fatalf("expected shop.id %d, got %v", shop.ID, shopObj["id"])
			}
		}
	})

	t.Run("public_list_filtered_by_multiple_shops", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/remessas-publicas?loja=%d&loja=%d", shop.ID, shop.ID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("consignment %d not found in multi-shop filtered listing: %s", cons.ID, rec.Body.String())
		}
	})

	t.Run("public_list_filtered_by_nonexistent_shop", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/remessas-publicas?loja=99999", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if len(consignments) != 0 {
			t.Fatalf("expected empty array for nonexistent shop, got %v", consignments)
		}
	})

	t.Run("public_list_by_shop_endpoint", func(t *testing.T) {
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

		path := fmt.Sprintf("/estabelecimentos/%s/remessas", shop.Slug)
		rec := h.doRequest(t, http.MethodGet, path, nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])

		if !containsID(consignments, cons.ID) {
			t.Fatalf("original consignment %d not found in shop listing: %s", cons.ID, rec.Body.String())
		}

		if containsID(consignments, otherCons.ID) {
			t.Fatalf("other shop's consignment %d should NOT appear in shop listing: %s", otherCons.ID, rec.Body.String())
		}

		otherPath := fmt.Sprintf("/estabelecimentos/%s/remessas", otherShop.Slug)
		otherRec := h.doRequest(t, http.MethodGet, otherPath, nil, "", nil)
		assertStatus(t, otherRec, http.StatusOK)

		otherBody := decodeJSON(t, otherRec)
		otherConsignments := asSlice(t, otherBody["data"])
		if !containsID(otherConsignments, otherCons.ID) {
			t.Fatalf("other shop's consignment %d should appear in its own listing: %s", otherCons.ID, otherRec.Body.String())
		}

		if containsID(otherConsignments, cons.ID) {
			t.Fatalf("original consignment %d should NOT appear in other shop listing: %s", cons.ID, otherRec.Body.String())
		}
	})

	// --- DELETE /remessas/:consignment/produtos/:product tests ---

	t.Run("remove_nonexistent_product_returns_404", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/remessas/%d/produtos/99999", cons.ID),
			&shopSession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("consignee_cannot_remove_product", func(t *testing.T) {
		monitorSession := h.login(t, monitor.Email, monitor.Phone)
		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/remessas/%d/produtos/%d", cons.ID, prd.ID),
			&monitorSession, "", nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("different_shop_cannot_remove_product", func(t *testing.T) {
		otherShopEmail := fmt.Sprintf("api-remove-other-%d@example.com", time.Now().UnixNano())
		_ = h.createShop(t, &admin,
			"Outra Loja RM", otherShopEmail, "11988887777")
		otherSession := h.login(t, otherShopEmail, "11988887777")

		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/remessas/%d/produtos/%d", cons.ID, prd.ID),
			&otherSession, "", nil)
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("remove_product_restores_stock", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/remessas/%d/produtos/%d", cons.ID, prd.ID),
			&shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if msg := asString(t, body["message"]); msg != "produto removido da remessa com sucesso" {
			t.Fatalf("expected success message, got %q", msg)
		}
	})

	t.Run("removed_product_not_in_consignment_listing", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/remessas/%d/produtos", cons.ID),
			&shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		if containsID(products, prd.ID) {
			t.Fatal("product should not be in consignment after removal")
		}
	})

	// --- status filtering: gated vs public ---

	t.Run("gated_list_includes_rascunho", func(t *testing.T) {
		h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"status": "RASCUNHO",
		})

		rec := h.doRequest(t, http.MethodGet, "/remessas", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("gated shop listing should include RASCUNHO consignment %d", cons.ID)
		}
	})

	t.Run("public_shop_list_excludes_rascunho", func(t *testing.T) {
		path := fmt.Sprintf("/estabelecimentos/%s/remessas", shop.Slug)
		rec := h.doRequest(t, http.MethodGet, path, nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if containsID(consignments, cons.ID) {
			t.Fatal("public shop listing should exclude RASCUNHO consignment")
		}
	})

	t.Run("public_list_excludes_rascunho", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/remessas-publicas", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if containsID(consignments, cons.ID) {
			t.Fatal("public listing should exclude RASCUNHO consignment")
		}
	})

	t.Run("gated_list_includes_arquivado", func(t *testing.T) {
		h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"status": "ARQUIVADO",
		})

		rec := h.doRequest(t, http.MethodGet, "/remessas", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("gated shop listing should include ARQUIVADO consignment %d", cons.ID)
		}
	})

	t.Run("public_shop_list_excludes_arquivado", func(t *testing.T) {
		path := fmt.Sprintf("/estabelecimentos/%s/remessas", shop.Slug)
		rec := h.doRequest(t, http.MethodGet, path, nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if containsID(consignments, cons.ID) {
			t.Fatal("public shop listing should exclude ARQUIVADO consignment")
		}
	})

	t.Run("public_list_excludes_arquivado", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/remessas-publicas", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if containsID(consignments, cons.ID) {
			t.Fatal("public listing should exclude ARQUIVADO consignment")
		}
	})

	// --- consignee gated listing: includes all statuses ---

	t.Run("consignee_gated_includes_rascunho", func(t *testing.T) {
		h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"status": "RASCUNHO",
		})

		monitorSession := h.login(t, monitor.Email, monitor.Phone)
		rec := h.doRequest(t, http.MethodGet, "/monitores/me/remessas", &monitorSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("consignee gated listing should include RASCUNHO consignment %d", cons.ID)
		}
	})

	t.Run("consignee_gated_includes_arquivado", func(t *testing.T) {
		h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"status": "ARQUIVADO",
		})

		monitorSession := h.login(t, monitor.Email, monitor.Phone)
		rec := h.doRequest(t, http.MethodGet, "/monitores/me/remessas", &monitorSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, cons.ID) {
			t.Fatalf("consignee gated listing should include ARQUIVADO consignment %d", cons.ID)
		}
	})

	// --- discount validation ---

	t.Run("discount_above_100_rejected", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": productName,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste discount",
					"category":    "porta-medalhas",
					"discount":    101,
					"featured":    false,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatal("expected discount=101 to be rejected")
		}
	})

	t.Run("negative_discount_rejected", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": productName,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste desconto negativo",
					"category":    "porta-medalhas",
					"discount":    -1,
					"featured":    false,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatal("expected discount=-1 to be rejected")
		}
	})

	// --- discountType validation ---

	t.Run("fixed_discount_valid", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-fdv-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("FDV %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste fixed discount",
					"category":    "porta-medalhas",
					"discount":    5,
					"discountType": "fixed",
					"featured":    false,
				},
			},
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("fixed_discount_greater_than_price", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-fdg-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("FDG %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste fixed discount > price",
					"category":    "porta-medalhas",
					"discount":    15,
					"discountType": "fixed",
					"featured":    false,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatal("expected fixed discount > price to be rejected")
		}
	})

	t.Run("fixed_discount_negative", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-fdn-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("FDN %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste fixed discount negative",
					"category":    "porta-medalhas",
					"discount":    -1,
					"discountType": "fixed",
					"featured":    false,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatal("expected negative fixed discount to be rejected")
		}
	})

	t.Run("explicit_percentage_discount", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-epd-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("EPD %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste explicit percentage",
					"category":    "porta-medalhas",
					"discount":    50,
					"discountType": "percentage",
					"featured":    false,
				},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify via GET — the POST response does not include product data
		rec2 := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		body := decodeJSON(t, rec2)
		products := asSlice(t, body["products"])
		for _, p := range products {
			pm := asMap(t, p)
			if asInt(t, pm["id"]) != prd.ID {
				continue
			}
			pivot := asMap(t, pm["pivot"])
			if dt := asString(t, pivot["discount_type"]); dt != "percentage" {
				t.Fatalf("expected discount_type 'percentage', got %q", dt)
			}
			return
		}
		t.Fatalf("product %d not found in consignment listing", prd.ID)
	})

	t.Run("percentage_discount_above_100", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-pda-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("PDA %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste percentage > 100",
					"category":    "porta-medalhas",
					"discount":    101,
					"discountType": "percentage",
					"featured":    false,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatal("expected percentage discount > 100 to be rejected")
		}
	})

	t.Run("percentage_discount_negative", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-pdn-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("PDN %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste percentage negative",
					"category":    "porta-medalhas",
					"discount":    -1,
					"discountType": "percentage",
					"featured":    false,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatal("expected negative percentage discount to be rejected")
		}
	})

	t.Run("invalid_discountType", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-idt-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("IDT %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste invalid discountType",
					"category":    "porta-medalhas",
					"discount":    10,
					"discountType": "bogus",
					"featured":    false,
				},
			},
		})
		if rec.Code == http.StatusOK {
			t.Fatal("expected invalid discountType to be rejected")
		}
	})

	t.Run("default_discountType_percentage", func(t *testing.T) {
		// omit discountType → defaults to "percentage"
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-ddp-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("DDP %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    1,
					"price":       10,
					"color":       "#000000",
					"description": "teste default discountType",
					"category":    "porta-medalhas",
					"discount":    20,
					"featured":    false,
				},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify via GET — the POST response does not include product data
		rec2 := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		body := decodeJSON(t, rec2)
		products := asSlice(t, body["products"])
		for _, p := range products {
			pm := asMap(t, p)
			if asInt(t, pm["id"]) != prd.ID {
				continue
			}
			pivot := asMap(t, pm["pivot"])
			if dt := asString(t, pivot["discount_type"]); dt != "percentage" {
				t.Fatalf("expected default discount_type 'percentage', got %q", dt)
			}
			return
		}
		t.Fatalf("product %d not found in consignment listing", prd.ID)
	})

	t.Run("update_consignment_product_discount", func(t *testing.T) {
		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-upd-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("UPD %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, prd.Name)

		// update discount + discountType
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d/produtos/%d", cons.ID, prd.ID), &shopSession, map[string]any{
			"discount":    3,
			"discountType": "fixed",
		})
		assertStatus(t, rec, http.StatusOK)

		// verify via GET
		rec2 := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		body := decodeJSON(t, rec2)
		products := asSlice(t, body["products"])
		for _, p := range products {
			pm := asMap(t, p)
			if asInt(t, pm["id"]) != prd.ID {
				continue
			}
			pivot := asMap(t, pm["pivot"])
			// discount is decimal.Decimal serialized as JSON number, captured as json.Number
			discountVal := fmt.Sprintf("%v", pivot["discount"])
			if discountVal != "3" && discountVal != "3.0" {
				t.Fatalf("expected discount 3, got %v", pivot["discount"])
			}
			if dt := asString(t, pivot["discount_type"]); dt != "fixed" {
				t.Fatalf("expected discount_type 'fixed', got %q", dt)
			}
			return
		}
		t.Fatalf("product %d not found in consignment listing", prd.ID)
	})

	t.Run("public_listing_returns_discount_type", func(t *testing.T) {
		// Reset consignment to LIBERADO so the public listing endpoint accepts it
		h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", cons.ID), &shopSession, map[string]any{
			"status": "LIBERADO",
		})

		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-pldt-%d", time.Now().UnixNano()))
		prd := h.createProduct(t, &shopSession, fmt.Sprintf("PLDT %d", time.Now().UnixNano()), cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		// Use explicit percentage to make the assertion unambiguous
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", cons.ID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   prd.ID,
					"productName": prd.Name,
					"quantity":    2,
					"price":       10,
					"color":       "#000000",
					"description": "teste public discount_type",
					"category":    "porta-medalhas",
					"discount":    15,
					"discountType": "percentage",
					"featured":    false,
				},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		code := h.getConsignmentCode(t, &shopSession, cons.ID)

		rec2 := h.doRequest(t, http.MethodGet, fmt.Sprintf("/produtos?token=%s", code), nil, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		body := decodeJSON(t, rec2)
		products := asSlice(t, body["produtos"])
		found := false
		for _, p := range products {
			pm := asMap(t, p)
			if asString(t, pm["name"]) != prd.Name {
				continue
			}
			found = true
			dt := asString(t, pm["discount_type"])
			if dt != "percentage" {
				t.Fatalf("expected discount_type 'percentage' in public listing, got %q", dt)
			}
			discountVal := fmt.Sprintf("%v", pm["discount"])
			if discountVal != "15" && discountVal != "15.0" {
				t.Fatalf("expected discount 15 in public listing, got %v", pm["discount"])
			}
		}
		if !found {
			t.Fatalf("product %q not found in public listing", prd.Name)
		}
	})
}
