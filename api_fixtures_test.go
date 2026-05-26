package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func uniqSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// --- fixture result types ---

type shopFixtureResult struct {
	ID   int
	Slug string
}

type monitorFixtureResult struct {
	ID    int
	Email string
	Phone string
}

type pickupFixtureResult struct {
	ID int
}

type categoryFixtureResult struct {
	ID   int
	Name string
}

type productFixtureResult struct {
	ID         int
	Name       string
	CategoryID int
	Stock      int
}

type consignmentFixtureResult struct {
	ID int
}

type saqueFixtureResult struct {
	ID          int
	Description string
}

type reportFixture struct {
	OrderID      any
	Monitor      string
	Moment       string
	Type         string
	Status       string
	Description  string
	Priority     string
	Scope        string
}

type reportFixtureResult struct {
	ID int
}

// --- fixture helpers ---

func (h *apiHarness) createShop(t *testing.T, session *apiSession, name, email, telephone string) shopFixtureResult {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/estabelecimentos", session, map[string]any{
		"name":      name,
		"email":     email,
		"telephone": telephone,
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
	return shopFixtureResult{
		ID:   asInt(t, shop["id"]),
		Slug: asString(t, shop["slug"]),
	}
}

func (h *apiHarness) createMonitor(t *testing.T, session *apiSession, name, email, phone string) monitorFixtureResult {
	t.Helper()

	uniquePhone := phone + "-" + uniqSuffix()
	rec := h.doJSON(t, http.MethodPost, "/monitores", session, map[string]any{
		"name":  name,
		"email": email,
		"phone": uniquePhone,
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
	return monitorFixtureResult{
		ID:    asInt(t, asMap(t, body["monitor"])["id"]),
		Email: email,
		Phone: uniquePhone,
	}
}

func (h *apiHarness) createPickup(t *testing.T, session *apiSession) pickupFixtureResult {
	t.Helper()

	pickupName := fmt.Sprintf("ponto-%s", uniqSuffix())
	rec := h.doJSON(t, http.MethodPost, "/pontos", session, map[string]any{
		"name":     pickupName,
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
	return pickupFixtureResult{
		ID: asInt(t, body["ponto"]),
	}
}

func (h *apiHarness) createCategory(t *testing.T, session *apiSession, name string) categoryFixtureResult {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/categorias", session, map[string]any{
		"name": name,
	})
	assertStatus(t, rec, http.StatusOK)

	body := decodeJSON(t, rec)
	return categoryFixtureResult{
		ID:   asInt(t, body["id"]),
		Name: name,
	}
}

func (h *apiHarness) createProduct(t *testing.T, session *apiSession, name string, categoryID int) productFixtureResult {
	t.Helper()

	categoryIDStr := fmt.Sprintf("%d", categoryID)
	rec := h.doMultipart(t, http.MethodPost, "/produtos", session, map[string]string{
		"name":                      name,
		"description":               "produto de teste",
		"price":                     "2",
		"category_id":               categoryIDStr,
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

	// Query shop products to get the product ID.
	rec = h.doRequest(t, http.MethodGet, "/produtos/estabelecimento", session, "", nil)
	assertStatus(t, rec, http.StatusOK)

	body := decodeJSON(t, rec)
	products := asSlice(t, body["data"])
	obj := findObjectByStringField(t, products, "name", name)
	id := asInt(t, obj["id"])
	stock := asInt(t, obj["stock"])

	return productFixtureResult{ID: id, Name: name, CategoryID: categoryID, Stock: stock}
}

func (h *apiHarness) approveProduct(t *testing.T, admin *apiSession, productID int) {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost,
		fmt.Sprintf("/admin/moderacao/produtos/%d/aprovar", productID),
		admin, map[string]any{
			"platform_cut":      5,
			"platform_cut_type": "percentage",
		})
	assertStatus(t, rec, http.StatusOK)
}

func (h *apiHarness) createConsignment(t *testing.T, session *apiSession, pickupID, monitorID int) consignmentFixtureResult {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/remessas", session, map[string]any{
		"destinationId": pickupID,
		"monitorId":     monitorID,
		"status":        "CONFERIR - DEVOLUCAO",
		"monitorProfit": 3,
	})
	assertStatus(t, rec, http.StatusOK)

	body := decodeJSON(t, rec)
	return consignmentFixtureResult{
		ID: asInt(t, body["id"]),
	}
}

func (h *apiHarness) addProductToConsignment(t *testing.T, session *apiSession, consignmentID, productID int, productName string) {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost,
		fmt.Sprintf("/remessas/%d/produtos", consignmentID),
		session, map[string]any{
			"products": []map[string]any{
				{
					"productId":   productID,
					"productName": productName,
					"quantity":    2,
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
}

func (h *apiHarness) createCartItem(t *testing.T, sessionID string, productID, quantity int) {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/carrinho/adicionar", nil, map[string]any{
		"product_id":      productID,
		"quantity":        quantity,
		"session_id":      sessionID,
		"personalization": map[string]any{},
	})
	assertStatus(t, rec, http.StatusOK)
}

func (h *apiHarness) getPackageID(t *testing.T, productID int, runID int64) int {
	t.Helper()

	sessionID := fmt.Sprintf("pkg-lookup-%d", runID)

	h.createCartItem(t, sessionID, productID, 1)

	optsRec := h.doRequest(t, http.MethodGet,
		fmt.Sprintf("/carrinho/pacotes?session_id=%s", sessionID),
		nil, "", nil)
	assertStatus(t, optsRec, http.StatusOK)

	body := decodeJSON(t, optsRec)
	products := asSlice(t, body["products"])
	if len(products) == 0 {
		t.Fatal("no products in package options")
	}

	pkg := findObjectByIntField(t, products, "product_id", productID)
	packages := asSlice(t, pkg["packages"])
	if len(packages) == 0 {
		t.Fatal("no packages in options")
	}

	return asInt(t, asMap(t, packages[0])["package_id"])
}

func (h *apiHarness) createSaque(t *testing.T, session *apiSession, description, askedValue, bankName string) saqueFixtureResult {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", session, map[string]any{
		"description": description,
		"asked_value": askedValue,
		"bank_name":   bankName,
		"pix_cpf":     "12345678900",
		"pix_titular": "Loja Teste",
	})
	assertStatus(t, rec, http.StatusOK)

	body := decodeJSON(t, rec)
	sq := asMap(t, body["saque"])
	return saqueFixtureResult{
		ID:          asInt(t, sq["id"]),
		Description: description,
	}
}

func (h *apiHarness) createReport(t *testing.T, session *apiSession, input reportFixture) reportFixtureResult {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/ocorrencias", session, map[string]any{
		"orderId":         input.OrderID,
		"monitor":         input.Monitor,
		"moment":          input.Moment,
		"type":            input.Type,
		"status":          input.Status,
		"description":     input.Description,
		"priority":        input.Priority,
		"occurrenceScope": input.Scope,
	})
	assertStatus(t, rec, http.StatusOK)

	body := decodeJSON(t, rec)
	return reportFixtureResult{ID: asInt(t, body["id"])}
}
