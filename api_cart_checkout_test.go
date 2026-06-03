package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCartAndCheckout(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Checkout %d", runID)
	shopEmail := fmt.Sprintf("api-checkout-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	monitor := h.createMonitor(t, &shopSession,
		fmt.Sprintf("Monitor C %d", runID),
		fmt.Sprintf("api-chk-monitor-%d@example.com", runID),
		"27998209673")
	pickup := h.createPickup(t, &shopSession)

	cat := h.createCategory(t, &shopSession, "lixo-eletronico")
	productName := fmt.Sprintf("Produto C %d", runID)
	prd := h.createProduct(t, &shopSession, productName, cat.ID)
	h.approveProduct(t, &admin, prd.ID)

	cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
	h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, productName)

	packageID := h.getPackageID(t, prd.ID, runID)

	_ = shop

	t.Run("add_product_auto_creates_cart", func(t *testing.T) {
		sessionID := fmt.Sprintf("auto-create-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 1)

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/dados?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		items := asSlice(t, body["items"])
		if len(items) != 1 {
			t.Fatalf("expected 1 item in cart, got %d: %s", len(items), rec.Body.String())
		}
	})

	t.Run("cart_view_missing_session", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/carrinho/dados", nil, "", nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("add_product_fails_missing_product_id", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/carrinho/adicionar", nil, map[string]any{
			"quantity":        1,
			"session_id":      fmt.Sprintf("no-product-%d", runID),
			"personalization": map[string]any{},
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("update_cart_quantity", func(t *testing.T) {
		sessionID := fmt.Sprintf("update-qty-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 1)

		viewRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/dados?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, viewRec, http.StatusOK)
		viewBody := decodeJSON(t, viewRec)
		cartItems := asSlice(t, viewBody["items"])
		if len(cartItems) != 1 {
			t.Fatalf("expected 1 cart item, got %d", len(cartItems))
		}
		updateItemID := asInt(t, asMap(t, cartItems[0])["id"])

		rec := h.doJSON(t, http.MethodPut,
			fmt.Sprintf("/carrinho/atualizar/%d?session_id=%s", updateItemID, sessionID),
			nil, map[string]any{"quantity": 5})
		assertStatus(t, rec, http.StatusOK)

		verifyRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/dados?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, verifyRec, http.StatusOK)
		verifyBody := decodeJSON(t, verifyRec)
		verifyItems := asSlice(t, verifyBody["items"])
		if len(verifyItems) != 1 {
			t.Fatalf("expected 1 cart item after update, got %d", len(verifyItems))
		}
		qty := asInt(t, asMap(t, verifyItems[0])["quantity"])
		if qty != 5 {
			t.Fatalf("expected quantity 5, got %d", qty)
		}
	})

	t.Run("remove_cart_item", func(t *testing.T) {
		sessionID := fmt.Sprintf("remove-item-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 2)

		viewRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/dados?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, viewRec, http.StatusOK)
		viewBody := decodeJSON(t, viewRec)
		cartItems := asSlice(t, viewBody["items"])
		if len(cartItems) != 1 {
			t.Fatalf("expected 1 cart item, got %d", len(cartItems))
		}
		removeItemID := asInt(t, asMap(t, cartItems[0])["id"])

		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/carrinho/remover/%d?session_id=%s", removeItemID, sessionID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		verifyRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/dados?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, verifyRec, http.StatusOK)
		verifyBody := decodeJSON(t, verifyRec)
		count := asInt(t, verifyBody["item_count"])
		if count != 0 {
			t.Fatal("expected cart item to be removed")
		}
	})

	t.Run("package_options", func(t *testing.T) {
		sessionID := fmt.Sprintf("test-cart-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 2)

		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/carrinho/pacotes?session_id=%s", sessionID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		if len(products) == 0 {
			t.Fatal("expected at least one product in response")
		}

		product := findObjectByIntField(t, products, "product_id", prd.ID)
		requireField(t, product, "product_name")

		packages := asSlice(t, product["packages"])
		if len(packages) == 0 {
			t.Fatal("expected at least one package option")
		}

		first := asMap(t, packages[0])
		if asInt(t, first["consignment_id"]) != cons.ID {
			t.Fatalf("expected consignment_id %d, got %v", cons.ID, first["consignment_id"])
		}

		for _, field := range []string{"price", "discount", "stock", "package_id"} {
			if _, ok := first[field]; !ok {
				t.Fatalf("expected %s field", field)
			}
		}

		pickupObj := asMap(t, first["pickup"])
		if asInt(t, pickupObj["id"]) != pickup.ID {
			t.Fatalf("expected pickup id %d, got %v", pickup.ID, pickupObj["id"])
		}
		if _, ok := pickupObj["name"]; !ok {
			t.Fatal("expected pickup name field")
		}
		if _, ok := pickupObj["date"]; !ok {
			t.Fatal("expected pickup date field")
		}

		address := asMap(t, pickupObj["address"])
		if asString(t, address["street"]) != "Rua Nova" {
			t.Fatalf("expected street 'Rua Nova', got %v", address["street"])
		}
	})

	t.Run("package_options_missing_session", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/carrinho/pacotes", nil, "", nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("package_options_empty_cart", func(t *testing.T) {
		sessionID := fmt.Sprintf("empty-cart-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 1)

		viewRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/dados?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, viewRec, http.StatusOK)
		viewBody := decodeJSON(t, viewRec)
		items := asSlice(t, viewBody["items"])
		if len(items) != 1 {
			t.Fatalf("expected 1 cart item, got %d", len(items))
		}
		emptyItemID := asInt(t, asMap(t, items[0])["id"])

		delRec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/carrinho/remover/%d?session_id=%s", emptyItemID, sessionID), nil, "", nil)
		assertStatus(t, delRec, http.StatusOK)

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/pacotes?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		if len(products) != 0 {
			t.Fatalf("expected empty products list, got %d items", len(products))
		}
	})

	t.Run("checkout_confirm_missing_session", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-NOSESSION-%d", runID),
			"items": []map[string]any{
				{"package_id": packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("checkout_confirm_missing_body", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost,
			"/checkouts/confirm?session_id=missing-body-"+fmt.Sprint(runID),
			nil, map[string]any{"value_paid": "3.00"})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("checkout_confirm", func(t *testing.T) {
		sessionID := fmt.Sprintf("checkout-test-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 1)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-%d", runID),
			"items": []map[string]any{
				{"package_id": packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if _, ok := body["unique_id"]; !ok {
			t.Fatal("expected unique_id in response")
		}
		if asString(t, body["status"]) != "paid" {
			t.Fatalf("expected status 'paid', got %v", body["status"])
		}

		qrCode, ok := body["qr_code"].(string)
		if !ok || qrCode == "" {
			t.Fatal("expected non-empty qr_code in response")
		}

		t.Run("idempotent_retry", func(t *testing.T) {
			rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
				"value_paid":         "3.00",
				"external_reference": fmt.Sprintf("EXT-REF-%d", runID),
				"items": []map[string]any{
					{"package_id": packageID, "quantity": 1},
				},
			})
			assertStatus(t, rec2, http.StatusOK)

			body2 := decodeJSON(t, rec2)
			if asString(t, body2["unique_id"]) != asString(t, body["unique_id"]) {
				t.Fatal("idempotent retry returned different unique_id")
			}
		})
	})

	t.Run("checkout_confirm_invalid_cart", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id=nonexistent-"+fmt.Sprint(runID), nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-INVALID-%d", runID),
			"items": []map[string]any{
				{"package_id": packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("checkout_confirm_empty_cart", func(t *testing.T) {
		sessionID := fmt.Sprintf("empty-checkout-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 1)

		viewRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/dados?session_id=%s", sessionID), nil, "", nil)
		assertStatus(t, viewRec, http.StatusOK)
		viewBody := decodeJSON(t, viewRec)
		emptyItems := asSlice(t, viewBody["items"])
		if len(emptyItems) != 1 {
			t.Fatalf("expected 1 cart item, got %d", len(emptyItems))
		}
		emptyItemID := asInt(t, asMap(t, emptyItems[0])["id"])

		delRec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/carrinho/remover/%d?session_id=%s", emptyItemID, sessionID), nil, "", nil)
		assertStatus(t, delRec, http.StatusOK)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-EMPTY-%d", runID),
			"items": []map[string]any{
				{"package_id": packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusUnprocessableEntity)
	})

	t.Run("checkout_confirm_insufficient_stock", func(t *testing.T) {
		sessionID := fmt.Sprintf("overstock-checkout-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 1)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-OVER-%d", runID),
			"items": []map[string]any{
				{"package_id": packageID, "quantity": 999},
			},
		})
		assertStatus(t, rec, http.StatusConflict)
	})

	t.Run("checkout_confirm_bad_package", func(t *testing.T) {
		sessionID := fmt.Sprintf("bad-pkg-%d", runID)

		h.createCartItem(t, sessionID, prd.ID, 1)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-BADPKG-%d", runID),
			"items": []map[string]any{
				{"package_id": 99999, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusUnprocessableEntity)
	})

	t.Run("checkout_reserved_reflected_in_package_options", func(t *testing.T) {
		preSession := fmt.Sprintf("pre-checkout-pkg-stock-%d", runID)

		h.createCartItem(t, preSession, prd.ID, 1)

		optsRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/pacotes?session_id=%s", preSession), nil, "", nil)
		assertStatus(t, optsRec, http.StatusOK)
		optsBody := decodeJSON(t, optsRec)
		preProduct := findObjectByIntField(t, asSlice(t, optsBody["products"]), "product_id", prd.ID)
		prePackages := asSlice(t, preProduct["packages"])
		if len(prePackages) == 0 {
			t.Fatal("expected at least one package in pre-checkout options")
		}
		preStock := asInt(t, asMap(t, prePackages[0])["stock"])

		checkoutSession := fmt.Sprintf("pkg-stock-checkout-%d", runID)
		h.createCartItem(t, checkoutSession, prd.ID, 1)

		chkRec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+checkoutSession, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-STOCK-%d", runID),
			"items": []map[string]any{
				{"package_id": packageID, "quantity": 1},
			},
		})
		assertStatus(t, chkRec, http.StatusOK)

		postSession := fmt.Sprintf("post-checkout-pkg-stock-%d", runID)
		h.createCartItem(t, postSession, prd.ID, 1)

		postOptsRec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/carrinho/pacotes?session_id=%s", postSession), nil, "", nil)
		assertStatus(t, postOptsRec, http.StatusOK)
		postOptsBody := decodeJSON(t, postOptsRec)
		postProduct := findObjectByIntField(t, asSlice(t, postOptsBody["products"]), "product_id", prd.ID)

		if preStock == 1 {
			// Package was already bought down to 1; this checkout sells the last unit.
			// With the refactored AvailablePackages, sold-out packages (quantity=0)
			// are excluded entirely via pkg.quantity > 0.
			if len(asSlice(t, postProduct["packages"])) != 0 {
				t.Fatal("expected no packages after last unit was sold")
			}
		} else {
			postPackages := asSlice(t, postProduct["packages"])
			if len(postPackages) == 0 {
				t.Fatal("expected at least one package in post-checkout options")
			}
			postStock := asInt(t, asMap(t, postPackages[0])["stock"])
			if postStock != preStock-1 {
				t.Fatalf("expected package stock to decrease by 1 after checkout (pre=%d, post=%d)", preStock, postStock)
			}
		}
	})
}
