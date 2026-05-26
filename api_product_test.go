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

	t.Run("list_all_public_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/produtos", nil, "", nil)
		assertStatus(t, rec, http.StatusOK)
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

	t.Run("delete_product", func(t *testing.T) {
		deleteName := fmt.Sprintf("Para Deletar %d", runID)
		deletePrd := h.createProduct(t, &shopSession, deleteName, cat.ID)

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/produtos/%d", deletePrd.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})
}
