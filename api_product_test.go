package main

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"
)

func TestProduct(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Produto %d", runID)
	shopEmail := fmt.Sprintf("api-prod-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	cat := h.createCategory(t, &shopSession, "lixo-eletronico")
	productName := fmt.Sprintf("Produto %d", runID)
	updatedName := fmt.Sprintf("Produto Atualizado %d", runID)

	prd := h.createProduct(t, &shopSession, productName, cat.ID)

	t.Run("create_product", func(t *testing.T) {
		if prd.ID == 0 {
			t.Fatal("expected non-zero product ID")
		}
		if prd.Stock <= 0 {
			t.Fatalf("expected stock > 0, got %d", prd.Stock)
		}
	})

	t.Run("list_shop_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/produtos/estabelecimento", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["data"])
		if len(products) < 1 {
			t.Fatal("expected shop to contain at least one product")
		}

		obj := findObjectByIntField(t, products, "id", prd.ID)
		listStock := asInt(t, obj["stock"])
		if listStock != prd.Stock {
			t.Fatalf("list endpoint stock %d does not match registered stock %d", listStock, prd.Stock)
		}
	})

	t.Run("update_product_form", func(t *testing.T) {
		// TODO: should be a fixture
		rec := h.doMultipart(t, http.MethodPut, fmt.Sprintf("/produtos/%d", prd.ID), &shopSession, map[string]string{
			"name":                      updatedName,
			"description":               "produto atualizado",
			"price":                     "0.11",
			"category_id":               fmt.Sprintf("%d", cat.ID),
			"stock":                     "10",
			"discount":                  "10",
			"allowCustomColorSelection": "false",
		}, []uploadFile{
			{
				Field:    "imageUrl",
				Filename: "product-updated.png",
				Data:     png1x1,
			},
			{
				Field:    "descriptionImages[]",
				Filename: "product-updated-desc.png",
				Data:     png1x1,
			},
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		data := asMap(t, body["data"])
		if asInt(t, data["id"]) != prd.ID {
			t.Fatalf("expected product id %d in response, got body %s", prd.ID, rec.Body.String())
		}
	})

	t.Run("pending_moderation_list", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/admin/moderacao/pendentes", &admin, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		pending := asSlice(t, body["data"])
		if !containsID(pending, prd.ID) {
			t.Fatalf("product %d not found in pending list: %s", prd.ID, rec.Body.String())
		}
	})

	t.Run("approve_product", func(t *testing.T) {
		h.approveProduct(t, &admin, prd.ID)
	})

	t.Run("public_shop_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/%s/produtos", shop.Slug), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		if !containsID(products, prd.ID) {
			t.Fatalf("product %d not found in public shop listing: %s", prd.ID, rec.Body.String())
		}
	})

	t.Run("list_all_public_products_without_token", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/produtos", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		produtos := asSlice(t, body["produtos"])
		retiradas := asSlice(t, body["retiradas"])
		if len(produtos) != 0 || len(retiradas) != 0 {
			t.Fatal("expected empty produtos and retiradas when no token")
		}
	})

	t.Run("view_public_product", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/produtos/%d", prd.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		product := asMap(t, body["data"])
		if asInt(t, product["id"]) != prd.ID {
			t.Fatalf("expected product id %d, got body %s", prd.ID, rec.Body.String())
		}

		viewStock := asInt(t, product["stock"])
		/*
			Reasoning: in the public view, the product stock
			MUST be calculated by the sum of stock in all
			the consignments. In this case, it must be 0
			because this product was not assigned to any
			consignment.
		*/
		if viewStock != 0 {
			t.Fatalf("view endpoint stock %d doesn't match sum of consignments (0)", viewStock)
		}

		category := asString(t, product["category"])
		if category != "lixo-eletronico" {
			t.Fatalf("expected category %q, got %q", "lixo-eletronico", category)
		}

		productCategoryID := asInt(t, product["category_id"])
		if productCategoryID != cat.ID {
			t.Fatalf("expected category ID %d, got %d", cat.ID, productCategoryID)
		}

		descImages := asSlice(t, product["descriptionImages"])
		if len(descImages) < 1 {
			t.Fatalf("expected product %d to have at least one description image", prd.ID)
		}
		img := asString(t, descImages[0])
		noProto := strings.TrimPrefix(img, "http://")
		base := path.Base(noProto)

		t.Run("delete_description_image", func(t *testing.T) {
			rec := h.doRequest(t, http.MethodDelete,
				fmt.Sprintf("/produtos/%d/img/%s", prd.ID, base),
				&shopSession, "", nil)
			assertStatus(t, rec, http.StatusOK)
		})
	})

	t.Run("list_by_token_invalid_uuid", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/produtos?token=not-a-uuid", nil, "", nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("list_by_token_not_found", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/produtos?token=00000000-0000-0000-0000-000000000000", nil, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("list_by_token_archived", func(t *testing.T) {
		monitorName := fmt.Sprintf("Monitor Arquivado %d", runID)
		monitorEmail := fmt.Sprintf("api-prod-arch-%d@example.com", runID)
		monitor := h.createMonitor(t, &shopSession, monitorName, monitorEmail, "27998209673")
		pickup := h.createPickup(t, &shopSession)

		cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
		code := h.getConsignmentCode(t, &shopSession, cons.ID)
		h.archiveConsignment(t, &shopSession, cons.ID)

		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/produtos?token=%s", code),
			nil, "", nil)
		assertStatus(t, rec, 410)
	})

	t.Run("list_by_token_success", func(t *testing.T) {
		monitorName := fmt.Sprintf("Monitor Token %d", runID)
		monitorEmail := fmt.Sprintf("api-prod-token-%d@example.com", runID)
		monitor := h.createMonitor(t, &shopSession, monitorName, monitorEmail, "27998209673")
		pickup := h.createPickup(t, &shopSession)

		cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
		h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, productName)
		code := h.getConsignmentCode(t, &shopSession, cons.ID)

		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/produtos?token=%s", code),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)

		// Check root-level retiradas
		retiradas := asSlice(t, body["retiradas"])
		if len(retiradas) == 0 {
			t.Fatal("expected non-empty retiradas in root level")
		}
		ret := asMap(t, retiradas[0])
		if asString(t, ret["name"]) == "" {
			t.Fatal("expected retirada name")
		}
		addr := asMap(t, ret["address"])
		for _, field := range []string{"street", "number", "neighborhood", "city", "state", "cep"} {
			if _, ok := addr[field]; !ok {
				t.Fatalf("expected address field %q in retirada", field)
			}
		}

		// Check produtos
		produtos := asSlice(t, body["produtos"])
		if len(produtos) == 0 {
			t.Fatal("expected non-empty produtos when listing by token")
		}
		prd := asMap(t, produtos[0])
		// id should be package_id
		if asInt(t, prd["id"]) == 0 {
			t.Fatal("expected non-zero id (package_id)")
		}
		// Should NOT have internal fields
		for _, field := range []string{"platform_cut", "platform_cut_type", "moderation_status"} {
			if _, ok := prd[field]; ok {
				t.Fatalf("unexpected internal field %q in product response", field)
			}
		}
		// Should NOT have pickups per product
		if _, ok := prd["pickups"]; ok {
			t.Fatal("expected no pickups field per product when listing by token")
		}
		// Should have consignment-specific fields
		for _, field := range []string{"name", "stock", "price", "discount", "featured"} {
			if _, ok := prd[field]; !ok {
				t.Fatalf("expected field %q in product", field)
			}
		}
	})

	t.Run("pickups_on_product_detail", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/produtos/%d", prd.ID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		product := asMap(t, body["data"])
		pickups := asSlice(t, product["pickups"])
		if len(pickups) == 0 {
			t.Fatal("expected non-empty pickups on product detail")
		}

		pkg := asMap(t, pickups[0])
		expectedFields := []string{"package_id", "price", "discount", "quantity", "pickup_name", "pickup_date", "address"}
		for _, field := range expectedFields {
			if _, ok := pkg[field]; !ok {
				t.Fatalf("expected field %q in pickup entry", field)
			}
		}

		addr := asMap(t, pkg["address"])
		addrFields := []string{"street", "number", "neighborhood", "city", "state", "cep"}
		for _, field := range addrFields {
			if _, ok := addr[field]; !ok {
				t.Fatalf("expected address field %q in pickup entry", field)
			}
		}
	})

	t.Run("empty_pickups_without_consignment", func(t *testing.T) {
		noConsignName := fmt.Sprintf("SemConsignacao %d", runID)
		noConsignPrd := h.createProduct(t, &shopSession, noConsignName, cat.ID)
		h.approveProduct(t, &admin, noConsignPrd.ID)

		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/produtos/%d", noConsignPrd.ID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		product := asMap(t, body["data"])
		pickups := asSlice(t, product["pickups"])
		if len(pickups) != 0 {
			t.Fatal("expected empty pickups array for product without consignments")
		}
	})

	t.Run("no_pickups_on_shop_listed_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/estabelecimentos/%s/produtos", shop.Slug),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		for _, p := range products {
			pm := asMap(t, p)
			if _, exists := pm["pickups"]; exists {
				t.Fatal("expected no pickups field on shop-listed products")
			}
		}
	})

	t.Run("consignments_field_removed", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/produtos/%d", prd.ID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		product := asMap(t, body["data"])
		if _, ok := product["consignments"]; ok {
			t.Fatal("expected no consignments field in product response")
		}
	})

	t.Run("delete_product", func(t *testing.T) {
		deleteName := fmt.Sprintf("Para Deletar %d", runID)
		deletePrd := h.createProduct(t, &shopSession, deleteName, cat.ID)

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/produtos/%d", deletePrd.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})
}
