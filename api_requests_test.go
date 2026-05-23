package main


import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/middleware/auth"
	"linhazero.com/pvapi/middleware/logmw"
	"linhazero.com/pvapi/seed"
)

type apiSession struct {
	cookies []*http.Cookie
}

type uploadFile struct {
	Field    string
	Filename string
	Data     []byte
}

type apiHarness struct {
	router *gin.Engine
	pool   *pgxpool.Pool
}

func TestAPIRequestsImplemented(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	registerEmail := fmt.Sprintf("api-user-%d@example.com", runID)
	shopEmail := fmt.Sprintf("api-shop-%d@example.com", runID)
	consigneeEmail := fmt.Sprintf("api-monitor-%d@example.com", runID)

	registerPassword := "abc123"
	adminEmail := "sistemas@lzero.org"
	adminPassword := "TestingPassword"
	shopPassword := "11988887777"
	consigneePassword := "27998209673"

	shopName := fmt.Sprintf("Nova Loja %d", runID)
	updatedShopName := fmt.Sprintf("Nova Loja Atualizada %d", runID)
	consigneeName := fmt.Sprintf("Pedro %d", runID)
	productName := fmt.Sprintf("Produto %d", runID)
	updatedProductName := fmt.Sprintf("Produto Atualizado %d", runID)

	var (
		adminSession     apiSession
		genericSession   apiSession
		shopSession      apiSession
		consigneeSession apiSession
		shopID         int
		shopSlug       string
		consigneeID    int
		pickupID       int
		categoryID     int
		productID      int
		consignmentID  int
		reportID       int
	)

	t.Run("user_no_auth_test", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/user", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("register", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/register", nil, map[string]any{
			"name":     "api user",
			"email":    registerEmail,
			"password": registerPassword,
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("login", func(t *testing.T) {
		genericSession = h.login(t, registerEmail, registerPassword)
	})

	t.Run("logout", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/logout", &genericSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("public_permissions", func(t *testing.T) {
		rec := h.doRequest(
			t,
			http.MethodGet,
			fmt.Sprintf("/permissions?email=%s",
				registerEmail,
			), nil, "", nil)

		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		report := asMap(t, body["data"])

		if report["highest"] != "user" {
			t.Fatalf("expected highest 'user', got %v",
				report["highest"])
		}

		rec = h.doRequest(t, http.MethodGet, fmt.Sprintf("/permissions?email=%s", adminEmail), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		report = asMap(t, body["data"])

		if report["highest"] != "admin" {
			t.Fatalf("expected highest 'admin', got %v",
				body["highest"])
		}
		if _, exists := report["admin"]; !exists {
			t.Fatal("expected 'admin' role to exist")
		}
	})

	adminSession = h.login(t, adminEmail, adminPassword)

	t.Run("create_shop_for_deactivation", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos", &adminSession, map[string]any{
			"name":      "inactive-co",
			"email":     "creative@ideas.com",
			"telephone": "bigone",
			"status":    "active",
			"address": map[string]any{
				"street":       "Rua Nova",
				"number":       "51",
				"neighborhood": "Bairro",
				"city":         "Campinas",
				"state":        "SP",
				"cep":          "13000000",
				"complement":   "Opcional",
			},
		})
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		shop := asMap(t, body["shop"])
		shopID := asInt(t, shop["id"])

		t.Run("deactivate_shop", func(t *testing.T) {
			rec := h.doRequest(
				t, http.MethodDelete,
				fmt.Sprintf("/estabelecimentos/%d", shopID),
				&adminSession,
				"", nil,
			)
			assertStatus(t, rec, http.StatusOK)
		})

		t.Run("list_admin_includes_all_shops", func(t *testing.T) {
			rec := h.doRequest(
				t, http.MethodGet,
				"/estabelecimentos/all",
				&adminSession,
				"", nil,
			)

			assertStatus(t, rec, http.StatusOK)

			body := decodeJSON(t, rec)
			shops := asSlice(t, body["shops"])

			if len(shops) == 0 {
				t.Fatal("no shops included in admin listing")
			}
		})

		t.Run("list_public_includes_no_shops", func(t *testing.T) {
			rec := h.doRequest(
				t, http.MethodGet,
				"/estabelecimentos",
				nil, "", nil,
			)

			assertStatus(t, rec, http.StatusOK)

			body := decodeJSON(t, rec)
			shops := asSlice(t, body["shops"])

			if len(shops) != 0 {
				t.Fatal("inactive shops included in public listing")
			}
		})
	})

	t.Run("create_shop", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/estabelecimentos", &adminSession, map[string]any{
			"name":      shopName,
			"email":     shopEmail,
			"telephone": shopPassword,
			"status":    "active",
			"address": map[string]any{
				"street":       "Rua Nova",
				"number":       "50",
				"neighborhood": "Bairro",
				"city":         "Campinas",
				"state":        "SP",
				"cep":          "13000000",
				"complement":   "Opcional",
			},
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		shop := asMap(t, body["shop"])
		shopID = asInt(t, shop["id"])
		shopSlug = asString(t, shop["slug"])
	})

	t.Run("list_shops", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos", &adminSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		shops := asSlice(t, body["shops"])
		if !containsID(shops, shopID) {
			t.Fatalf("shop %d not found in response: %s", shopID, rec.Body.String())
		}
	})

	t.Run("view_shop", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/%d", shopID), &adminSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		shop := asMap(t, body["shop"])
		if asInt(t, shop["id"]) != shopID {
			t.Fatalf("expected shop id %d, got body %s", shopID, rec.Body.String())
		}
	})

	t.Run("update_shop", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/estabelecimentos/%d", shopID), &adminSession, map[string]any{
			"name": updatedShopName,
		})
		assertStatus(t, rec, http.StatusOK)
	})

	shopSession = h.login(t, shopEmail, shopPassword)

	t.Run("create_monitor", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/monitores", &shopSession, map[string]any{
			"name":  consigneeName,
			"email": consigneeEmail,
			"phone": consigneePassword,
			"address": map[string]any{
				"street":       "Rua Nova",
				"number":       "50",
				"neighborhood": "Bairro",
				"city":         "Campinas",
				"state":        "SP",
				"cep":          "13000000",
			},
		})
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		consigneeID = asInt(t, asMap(t, body["monitor"])["id"])
	})

	t.Run("list_monitors", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/monitores", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		//body := decodeJSON(t, rec)
		//monitors := asSlice(t, body["data"])
		//consigneeID = findIDByFieldValue(t, monitors, "email", consigneeEmail)
	})

	t.Run("delete_monitor", func(t *testing.T) {
		deleteEmail := fmt.Sprintf("api-del-monitor-%d@example.com", runID)

		rec := h.doJSON(t, http.MethodPost, "/monitores", &shopSession, map[string]any{
			"name":  fmt.Sprintf("Para Deletar %d", runID),
			"email": deleteEmail,
			"phone": "11999999999",
			"address": map[string]any{
				"street":       "Rua Nova",
				"number":       "50",
				"neighborhood": "Bairro",
				"city":         "Campinas",
				"state":        "SP",
				"cep":          "13000000",
			},
		})
		assertStatus(t, rec, http.StatusOK)

		deleteID := asInt(t, asMap(t, decodeJSON(t, rec)["monitor"])["id"])

		rec = h.doRequest(t, http.MethodDelete, fmt.Sprintf("/monitores/%d", deleteID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("monitor_login", func(t *testing.T) {
		consigneeSession = h.login(t, consigneeEmail, consigneePassword)
	})

	t.Run("create_pickup_location", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/pontos", &shopSession, map[string]any{
			"name":     "ponto",
			"datetime": "2024-03-25T14:05:43",
			"address": map[string]any{
				"street":       "Rua Nova",
				"number":       "50",
				"neighborhood": "Bairro",
				"city":         "Campinas",
				"state":        "SP",
				"cep":          "13000000",
				"complement":   "Opcional",
			},
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		pickupID = asInt(t, body["ponto"])
	})

	t.Run("list_pickup_locations", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/pontos", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		pickups := asSlice(t, body["data"])
		if !containsID(pickups, pickupID) {
			t.Fatalf("pickup %d not found in response: %s", pickupID, rec.Body.String())
		}
	})

	t.Run("create_category", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/categorias", &shopSession, map[string]any{
			"name": "lixo-eletronico",
		})
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		categoryID = asInt(t, body["id"])
	})

	t.Run("list_categories", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/categorias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		categories := asSlice(t, body["data"])
		if len(categories) < 1 {
			t.Fatal("expected at least one category in list.")
		}
		// first one is n/a
		matched := asMap(t, categories[1])
		matchedID := asInt(t, matched["id"])
		if categoryID != matchedID {
			t.Fatalf("Category id (%d), is not equal to listed id (%d).",
				categoryID, matchedID,
			)
		}
	})

	t.Run("create_product_for_deletion", func(t *testing.T) {
		rec := h.doMultipart(t, http.MethodPost, "/produtos", &shopSession, map[string]string{
			"name":                      productName,
			"description":               "produto de teste temporario",
			"price":                     "2",
			"category_id":               "2",
			"stock":                     "10",
			"featured":                  "true",
			"discount":                  "12",
			"allowCustomName":           "false",
			"allowCustomModality":       "false",
			"allowCustomColorSelection": "false",
			"colors":                    `["#000000"]`,
			"specifications":            `[{"name":"material","value":"metal"}]`,
		}, []uploadFile{
			{
				Field:    "imageUrl",
				Filename: "product.png",
				Data:     png1x1,
			},
			{
				Field:    "descriptionImages[]",
				Filename: "product.png",
				Data:     png1x1,
			},
		})
		assertStatus(t, rec, http.StatusOK)

		t.Run("delete_product", func(t *testing.T) {
			rec := h.doRequest(t, http.MethodDelete, "/produtos/1", &shopSession, "", nil)
			assertStatus(t, rec, http.StatusOK)
		})
	})

	t.Run("create_product", func(t *testing.T) {
		rec := h.doMultipart(t, http.MethodPost, "/produtos", &shopSession, map[string]string{
			"name":                      productName,
			"description":               "produto de teste",
			"price":                     "2",
			"category_id":               "1",
			"stock":                     "10",
			"featured":                  "true",
			"discount":                  "12",
			"allowCustomName":           "false",
			"allowCustomModality":       "false",
			"allowCustomColorSelection": "false",
			"colors":                    `["#000000"]`,
			"specifications":            `[{"name":"material","value":"metal"}]`,
		}, []uploadFile{
			{
				Field:    "imageUrl",
				Filename: "product.png",
				Data:     png1x1,
			},
			{
				Field:    "descriptionImages[]",
				Filename: "product.png",
				Data:     png1x1,
			},
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("list_shop_owner_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/produtos/estabelecimento", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["data"])
		productID = findIDByFieldValue(t, products, "name", productName)
		if len(products) < 1 {
			t.Fatal("expected shop to contain at least one product")
		}
		product := asMap(t, products[0])
		category := asString(t, product["category"])
		if category != "n/a" {
			t.Fatalf("product category is wrong, expected %q, got %q", "n/a", category)
		}
	})

	t.Run("update_product_form", func(t *testing.T) {
		rec := h.doMultipart(t, http.MethodPut, fmt.Sprintf("/produtos/%d", productID), &shopSession, map[string]string{
			"name":                      updatedProductName,
			"description":               "produto atualizado",
			"price":                     "0.11",
			"category_id":               fmt.Sprintf("%d", categoryID),
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
	})

	t.Run("view_pending_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/admin/moderacao/pendentes", &adminSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		pending := asSlice(t, body["data"])
		if !containsID(pending, productID) {
			t.Fatalf("product %d not found in pending list: %s", productID, rec.Body.String())
		}
	})

	t.Run("approve_product", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/admin/moderacao/produtos/%d/aprovar", productID), &adminSession, map[string]any{
			"platform_cut":      5,
			"platform_cut_type": "percentage",
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("list_shop_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/estabelecimentos/%s/produtos", shopSlug), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		if !containsID(products, productID) {
			t.Fatalf("product %d not found in public shop listing: %s", productID, rec.Body.String())
		}
	})

	t.Run("list_all_public_products", func(t *testing.T) {
		rec := h.doRequest(
			t,
			http.MethodGet,
			"/produtos",
			nil, "", nil,
		)

		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("view_public_product", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/produtos/%d", productID), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		product := asMap(t, body["data"])
		if asInt(t, product["id"]) != productID {
			t.Fatalf("expected product id %d, got body %s", productID, rec.Body.String())
		}

		category := asString(t, product["category"])
		if category != "lixo-eletronico" {
			t.Fatalf("expected product %q's category to be %q, got %q",
				productID, "lixo-eletronico", category)
		}

		productCategoryID := asInt(t, product["category_id"])
		if productCategoryID != categoryID {
			t.Fatalf("expected product %q's category ID to be %q, got %q",
				productID, categoryID, productCategoryID)
		}

		descImages := asSlice(t, product["descriptionImages"])
		if len(descImages) < 1 {
			t.Fatalf("expected product %q to have at least one description image", productID)
		}
		img := asString(t, descImages[0])
		noProto := strings.TrimPrefix(img, "http://")
		base := path.Base(noProto)

		t.Run("delete_image", func(t *testing.T) {
			rec := h.doRequest(
				t, http.MethodDelete,
				fmt.Sprintf(
					"/produtos/%d/img/%s",
					productID, base,
				),
				&shopSession,
				"",
				nil,
			)
			assertStatus(t, rec, http.StatusOK)
		})
	})

	t.Run("create_consignment", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/remessas", &shopSession, map[string]any{
			"destinationId": pickupID,
			"monitorId":     consigneeID,
			"status":        "RASCUNHO",
			"monitorProfit": 3,
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignmentID = asInt(t, body["id"])
	})

	t.Run("list_consignments", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/remessas", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignments := asSlice(t, body["data"])
		if !containsID(consignments, consignmentID) {
			t.Fatalf("consignment %d not found in response: %s", consignmentID, rec.Body.String())
		}
	})

	t.Run("view_consignment", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d", consignmentID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		consignment := asMap(t, body["data"])
		if asInt(t, consignment["id"]) != consignmentID {
			t.Fatalf("expected consignment id %d, got body %s", consignmentID, rec.Body.String())
		}
	})

	t.Run("update_consignment", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/remessas/%d", consignmentID), &shopSession, map[string]any{
			"destinationId": pickupID,
			"monitorId":     consigneeID,
			"status":        "CONFERIR - DEVOLUCAO",
			"monitorProfit": 5,
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("add_product_to_consignment", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", consignmentID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   productID,
					"productName": updatedProductName,
					"quantity":    1,
					"price":       3,
					"color":       "#000000",
					"description": "produto atualizado",
					"category":    "porta-medalhas",
					"discount":    10,
					"featured":    true,
				},
			},
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("list_consignment_products", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/remessas/%d/produtos", consignmentID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])
		if !containsID(products, productID) {
			t.Fatalf("product %d not found in consignment response: %s", productID, rec.Body.String())
		}
	})

	t.Run("cart_consignment_options", func(t *testing.T) {
		sessionID := fmt.Sprintf("test-cart-%d", runID)

		_, err := h.pool.Exec(context.Background(),
			"INSERT INTO carts (session_id) VALUES ($1)", sessionID)
		if err != nil {
			t.Fatalf("create cart: %v", err)
		}

		_, err = h.pool.Exec(context.Background(),
			`INSERT INTO cart_product (product_id, quantity, personalizations, cart_id)
			 SELECT $1, 2, '{}', id FROM carts WHERE session_id = $2`,
			productID, sessionID)
		if err != nil {
			t.Fatalf("add product to cart: %v", err)
		}

		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/carrinho/remessas?session_id=%s", sessionID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])

		if len(products) == 0 {
			t.Fatal("expected at least one product in response")
		}

		product := asMap(t, products[0])
		if asInt(t, product["product_id"]) != productID {
			t.Fatalf("expected product_id %d, got %v", productID, product["product_id"])
		}

		if _, ok := product["product_name"]; !ok {
			t.Fatal("expected product_name field")
		}

		consignments := asSlice(t, product["consignments"])
		if len(consignments) == 0 {
			t.Fatal("expected at least one consignment option")
		}

		first := asMap(t, consignments[0])
		if asInt(t, first["consignment_id"]) != consignmentID {
			t.Fatalf("expected consignment_id %d, got %v", consignmentID, first["consignment_id"])
		}

		for _, field := range []string{"price_in_consignment", "discount", "consignment_stock"} {
			if _, ok := first[field]; !ok {
				t.Fatalf("expected %s field", field)
			}
		}

		pickup := asMap(t, first["pickup"])
		if asInt(t, pickup["id"]) != pickupID {
			t.Fatalf("expected pickup id %d, got %v", pickupID, pickup["id"])
		}

		if _, ok := pickup["name"]; !ok {
			t.Fatal("expected pickup name field")
		}
		if _, ok := pickup["date"]; !ok {
			t.Fatal("expected pickup date field")
		}

		address := asMap(t, pickup["address"])
		if asString(t, address["street"]) != "Rua Nova" {
			t.Fatalf("expected street 'Rua Nova', got %v", address["street"])
		}
	})

	t.Run("cart_consignment_options_missing_session", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/carrinho/remessas", nil, "", nil)
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("cart_consignment_options_empty_cart", func(t *testing.T) {
		sessionID := fmt.Sprintf("empty-cart-%d", runID)

		_, err := h.pool.Exec(context.Background(),
			"INSERT INTO carts (session_id) VALUES ($1)", sessionID)
		if err != nil {
			t.Fatalf("create cart: %v", err)
		}

		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/carrinho/remessas?session_id=%s", sessionID),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		products := asSlice(t, body["products"])

		if len(products) != 0 {
			t.Fatalf("expected empty products list, got %d items", len(products))
		}
	})

	t.Run("shouldFail_add_absurd_ammount_of_products", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, fmt.Sprintf("/remessas/%d/produtos", consignmentID), &shopSession, map[string]any{
			"products": []map[string]any{
				{
					"productId":   productID,
					"productName": updatedProductName,
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

	t.Run("create_report_nonexistent_product_fails", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         99999,
			"monitor":         consigneeName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "product",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_report_nonexistent_consignment_fails", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         99999,
			"monitor":         consigneeName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "consignment",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_report_unknown_scope_fails", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         productID,
			"monitor":         consigneeName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "invalid",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_report_product_scope", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         productID,
			"monitor":         consigneeName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "product",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		reportID = asInt(t, body["id"])
	})

	t.Run("list_reports", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/ocorrencias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		reports := asSlice(t, body["data"])
		if !containsID(reports, reportID) {
			t.Fatalf("report %d not found in response: %s", reportID, rec.Body.String())
		}
	})

	t.Run("view_report", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/ocorrencias/%d", reportID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asInt(t, report["id"]) != reportID {
			t.Fatalf("expected report id %d, got body %s", reportID, rec.Body.String())
		}
	})

	t.Run("consignee_view_product_report", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/ocorrencias/%d", reportID), &consigneeSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asInt(t, report["id"]) != reportID {
			t.Fatalf("expected report id %d, got body %s", reportID, rec.Body.String())
		}
	})

	t.Run("consignee_list_reports_includes_product_report", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/ocorrencias", &consigneeSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		reports := asSlice(t, body["data"])
		if !containsID(reports, reportID) {
			t.Fatalf("report %d not found in consignee report list: %s", reportID, rec.Body.String())
		}
	})

	t.Run("delete_report", func(t *testing.T) {
		rec := h.doRequest(
			t,
			http.MethodDelete,
			fmt.Sprintf("/ocorrencias/%d", reportID),
			&shopSession,
			"", nil,
		)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("checkout_confirm", func(t *testing.T) {
		sessionID := fmt.Sprintf("checkout-test-%d", runID)

		_, err := h.pool.Exec(context.Background(),
			"INSERT INTO carts (session_id) VALUES ($1)", sessionID)
		if err != nil {
			t.Fatalf("create cart: %v", err)
		}

		var cartID int
		err = h.pool.QueryRow(context.Background(),
			"SELECT id FROM carts WHERE session_id = $1", sessionID).Scan(&cartID)
		if err != nil {
			t.Fatalf("get cart id: %v", err)
		}

		_, err = h.pool.Exec(context.Background(),
			`INSERT INTO cart_product (product_id, quantity, personalizations, cart_id)
			 VALUES ($1, 1, '{}', $2)`,
			productID, cartID)
		if err != nil {
			t.Fatalf("add product to cart: %v", err)
		}

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-%d", runID),
			"items": []map[string]any{
				{
					"product_id":     productID,
					"consignment_id": consignmentID,
					"quantity":       1,
				},
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

		t.Run("idempotent_retry", func(t *testing.T) {
			rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
				"value_paid":         "3.00",
				"external_reference": fmt.Sprintf("EXT-REF-%d", runID),
				"items": []map[string]any{
					{
						"product_id":     productID,
						"consignment_id": consignmentID,
						"quantity":       1,
					},
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
				{
					"product_id":     productID,
					"consignment_id": consignmentID,
					"quantity":       1,
				},
			},
		})
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("checkout_confirm_empty_cart", func(t *testing.T) {
		sessionID := fmt.Sprintf("empty-checkout-%d", runID)

		_, err := h.pool.Exec(context.Background(),
			"INSERT INTO carts (session_id) VALUES ($1)", sessionID)
		if err != nil {
			t.Fatalf("create empty cart: %v", err)
		}

		var emptyCartID int
		err = h.pool.QueryRow(context.Background(),
			"SELECT id FROM carts WHERE session_id = $1", sessionID).Scan(&emptyCartID)
		if err != nil {
			t.Fatalf("get empty cart id: %v", err)
		}

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-EMPTY-%d", runID),
			"items": []map[string]any{
				{
					"product_id":     productID,
					"consignment_id": consignmentID,
					"quantity":       1,
				},
			},
		})
		assertStatus(t, rec, http.StatusUnprocessableEntity)
	})

	t.Run("checkout_confirm_insufficient_stock", func(t *testing.T) {
		sessionID := fmt.Sprintf("overstock-checkout-%d", runID)

		_, err := h.pool.Exec(context.Background(),
			"INSERT INTO carts (session_id) VALUES ($1)", sessionID)
		if err != nil {
			t.Fatalf("create cart: %v", err)
		}

		var overCartID int
		err = h.pool.QueryRow(context.Background(),
			"SELECT id FROM carts WHERE session_id = $1", sessionID).Scan(&overCartID)
		if err != nil {
			t.Fatalf("get cart id: %v", err)
		}

		_, err = h.pool.Exec(context.Background(),
			`INSERT INTO cart_product (product_id, quantity, personalizations, cart_id)
			 VALUES ($1, 999, '{}', $2)`,
			productID, overCartID)
		if err != nil {
			t.Fatalf("add product to cart: %v", err)
		}

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-OVER-%d", runID),
			"items": []map[string]any{
				{
					"product_id":     productID,
					"consignment_id": consignmentID,
					"quantity":       999,
				},
			},
		})
		assertStatus(t, rec, http.StatusConflict)
	})

	t.Run("checkout_confirm_product_not_in_consignment", func(t *testing.T) {
		sessionID := fmt.Sprintf("wrong-consignment-%d", runID)

		_, err := h.pool.Exec(context.Background(),
			"INSERT INTO carts (session_id) VALUES ($1)", sessionID)
		if err != nil {
			t.Fatalf("create cart: %v", err)
		}

		var wrongCartID int
		err = h.pool.QueryRow(context.Background(),
			"SELECT id FROM carts WHERE session_id = $1", sessionID).Scan(&wrongCartID)
		if err != nil {
			t.Fatalf("get cart id: %v", err)
		}

		_, err = h.pool.Exec(context.Background(),
			`INSERT INTO cart_product (product_id, quantity, personalizations, cart_id)
			 VALUES ($1, 1, '{}', $2)`,
			productID, wrongCartID)
		if err != nil {
			t.Fatalf("add product to cart: %v", err)
		}

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-WRONG-%d", runID),
			"items": []map[string]any{
				{
					"product_id":     productID,
					"consignment_id": 99999,
					"quantity":       1,
				},
			},
		})
		assertStatus(t, rec, http.StatusUnprocessableEntity)
	})

	t.Run("list_clients", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/clientes", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if _, ok := body["clients"]; !ok {
			t.Fatalf("expected 'clients' field in response, got %s", rec.Body.String())
		}
	})
}

func newAPIHarness(t *testing.T) *apiHarness {
	t.Helper()

	setDefaultEnv(t, "PV_API_SECURE_COOKIE", "no")
	setDefaultEnv(t, "PV_API_SECRET", "secret")
	setDefaultEnv(t, "PV_API_DB_HOST", "127.0.0.1")
	setDefaultEnv(t, "PV_API_DB_USER", "test_user")
	setDefaultEnv(t, "PV_API_DB_PASS", "testing_database")
	setDefaultEnv(t, "PV_API_DB_BASE", "test_db")
	setDefaultEnv(t, "PV_API_DB_PORT", "5434")
	setDefaultEnv(t, "PV_API_ADMIN_USER", "lzero")
	setDefaultEnv(t, "PV_API_ADMIN_MAIL", "sistemas@lzero.org")
	setDefaultEnv(t, "PV_API_ADMIN_PASS", "TestingPassword")
	setDefaultEnv(t, "PV_API_LOG_LEVEL", "debug")
	setDefaultEnv(t, "PV_API_DOMAIN_NAME", "localhost")
	setDefaultEnv(
		t,
		"PV_API_MP_TOKEN",
		"TEST-placeholder-mercadopago-token",
	)
	setDefaultEnv(t, "PV_API_TEST_MODE", "yes")

	if err := os.MkdirAll("public/uploads", 0o755); err != nil {
		t.Fatalf("create uploads dir: %v", err)
	}

	gin.SetMode(gin.TestMode)

	pool, err := db.NewPool()
	if err != nil {
		t.Skipf("test database not configured: %v", err)
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		t.Skipf("test database is not reachable; start it with test.just first: %v", err)
	}

	//out := os.Stdout; _=io.Discard
	out := io.Discard
	logger := slog.New(slog.NewTextHandler(out, nil))

	if err := seed.Plant(logger, context.Background(), pool); err != nil {
		pool.Close()
		t.Fatalf("seed admin user: %v", err)
	}

	params, err := auth.GenerateJwtInitParams(pool)
	if err != nil {
		pool.Close()
		t.Fatalf("generate auth params: %v", err)
	}

	authMiddleware, err := jwt.New(params)
	if err != nil {
		pool.Close()
		t.Fatalf("build auth middleware: %v", err)
	}

	if err := authMiddleware.MiddlewareInit(); err != nil {
		pool.Close()
		t.Fatalf("init auth middleware: %v", err)
	}

	router := gin.New()
	router.Use(logmw.LogMiddleware(logger))
	router.Use(gin.Recovery())
	setupRoutes(router, authMiddleware, pool)

	t.Cleanup(pool.Close)

	return &apiHarness{
		router: router,
		pool:   pool,
	}
}

func setDefaultEnv(t *testing.T, key, value string) {
	t.Helper()
	if _, ok := os.LookupEnv(key); !ok {
		t.Setenv(key, value)
	}
}

func (h *apiHarness) login(t *testing.T, email, password string) apiSession {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/login", nil, map[string]any{
		"email":    email,
		"password": password,
	})
	assertStatus(t, rec, http.StatusOK)

	return apiSession{
		cookies: rec.Result().Cookies(),
	}
}

func (h *apiHarness) doJSON(
	t *testing.T,
	method, path string,
	session *apiSession,
	payload any,
) *httptest.ResponseRecorder {
	t.Helper()

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(data)
	}

	return h.doRequest(t, method, path, session, "application/json", body)
}

func (h *apiHarness) doMultipart(
	t *testing.T,
	method, path string,
	session *apiSession,
	fields map[string]string,
	files []uploadFile,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write multipart field %q: %v", key, err)
		}
	}

	for _, file := range files {
		part, err := writer.CreateFormFile(file.Field, file.Filename)
		if err != nil {
			t.Fatalf("create multipart file %q: %v", file.Field, err)
		}
		if _, err := part.Write(file.Data); err != nil {
			t.Fatalf("write multipart file %q: %v", file.Field, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart body: %v", err)
	}

	return h.doRequest(t, method, path, session, writer.FormDataContentType(), &body)
}

func (h *apiHarness) doRequest(
	t *testing.T,
	method, path string,
	session *apiSession,
	contentType string,
	body io.Reader,
) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if session != nil {
		for _, cookie := range session.cookies {
			req.AddCookie(cookie)
		}
	}

	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)
	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("expected status %d, got %d: %s", want, rec.Code, rec.Body.String())
	}
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	dec := json.NewDecoder(bytes.NewReader(rec.Body.Bytes()))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		t.Fatalf("decode json response: %v; body=%s", err, rec.Body.String())
	}
	return body
}

func asMap(t *testing.T, value any) map[string]any {
	t.Helper()
	v, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", value)
	}
	return v
}

func asSlice(t *testing.T, value any) []any {
	t.Helper()
	v, ok := value.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", value)
	}
	return v
}

func asString(t *testing.T, value any) string {
	t.Helper()
	v, ok := value.(string)
	if !ok {
		t.Fatalf("expected string, got %T", value)
	}
	return v
}

func asInt(t *testing.T, value any) int {
	t.Helper()

	switch v := value.(type) {
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			t.Fatalf("convert json number to int: %v", err)
		}
		return int(i)
	case float64:
		return int(v)
	case int:
		return v
	default:
		t.Fatalf("expected number, got %T", value)
		return 0
	}
}

func containsID(items []any, want int) bool {
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, ok := obj["id"]
		if !ok {
			continue
		}
		switch v := id.(type) {
		case json.Number:
			n, err := v.Int64()
			if err == nil && int(n) == want {
				return true
			}
		case float64:
			if int(v) == want {
				return true
			}
		case string:
			if string(v) == strconv.Itoa(want) {
				return true
			}
		}
	}
	return false
}

func findIDByFieldValue(t *testing.T, items []any, field, want string) int {
	t.Helper()

	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		got, ok := obj[field].(string)
		if !ok || got != want {
			continue
		}

		return asInt(t, obj["id"])
	}

	t.Fatalf("could not find item where %s=%q", field, want)
	return 0
}

var png1x1 = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x04, 0x00, 0x00, 0x00, 0xb5, 0x1c, 0x0c,
	0x02, 0x00, 0x00, 0x00, 0x0b, 0x49, 0x44, 0x41,
	0x54, 0x78, 0xda, 0x63, 0xfc, 0xff, 0x1f, 0x00,
	0x03, 0x03, 0x02, 0x00, 0xef, 0x45, 0xca, 0x95,
	0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
	0xae, 0x42, 0x60, 0x82,
}
