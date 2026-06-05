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

type bankAccountFixtureResult struct {
	ID int
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
		"destinationId": []any{pickupID},
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

func (h *apiHarness) getPackageID(t *testing.T, productID int, runID int64) int {
	t.Helper()

	rec := h.doRequest(t, http.MethodGet,
		fmt.Sprintf("/produtos/%d", productID),
		nil, "", nil)
	assertStatus(t, rec, http.StatusOK)

	body := decodeJSON(t, rec)
	product := asMap(t, body["data"])
	pickups := asSlice(t, product["pickups"])
	if len(pickups) == 0 {
		t.Fatal("no pickups available for product")
	}

	return asInt(t, asMap(t, pickups[0])["package_id"])
}

func (h *apiHarness) createBankAccount(t *testing.T, session *apiSession, name, bankName string) bankAccountFixtureResult {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/contasbancarias", session, map[string]any{
		"name":      name,
		"bank_name": bankName,
		"pix_key":   "12345678900",
		"pix_titular": "Loja Teste",
	})
	assertStatus(t, rec, http.StatusOK)

	body := decodeJSON(t, rec)
	conta := asMap(t, body["conta"])
	return bankAccountFixtureResult{
		ID: asInt(t, conta["id"]),
	}
}

func (h *apiHarness) generateEarnings(t *testing.T, session *apiSession, admin *apiSession, shop shopFixtureResult, runID int64) int {
	t.Helper()

	monitor := h.createMonitor(t, session,
		fmt.Sprintf("Monitor E %d", runID),
		fmt.Sprintf("api-earnings-monitor-%d@example.com", runID),
		"27998209673")
	pickup := h.createPickup(t, session)

	cat := h.createCategory(t, session, fmt.Sprintf("cat-earnings-%d", runID))
	productName := fmt.Sprintf("Produto E %d", runID)
	prd := h.createProduct(t, session, productName, cat.ID)
	h.approveProduct(t, admin, prd.ID)

	// Use PERCENT comission so net earnings are positive
	rec := h.doJSON(t, http.MethodPost, "/remessas", session, map[string]any{
		"destinationId":           []any{pickup.ID},
		"monitorId":               monitor.ID,
		"status":                  "CONFERIR - DEVOLUCAO",
		"monitorProfitPercentage": 1,
	})
	assertStatus(t, rec, http.StatusOK)
	body := decodeJSON(t, rec)
	cons := consignmentFixtureResult{ID: asInt(t, body["id"])}

	h.addProductToConsignment(t, session, cons.ID, prd.ID, productName)

	packageID := h.getPackageID(t, prd.ID, runID)


	chkRec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
		"value_paid":         "3.00",
		"external_reference": fmt.Sprintf("EXT-REF-EARN-%d", runID),
		"items": []map[string]any{
			{"package_id": packageID, "quantity": 1},
		},
	})
	assertStatus(t, chkRec, http.StatusOK)

	return packageID
}

func (h *apiHarness) createSaque(t *testing.T, session *apiSession, description, askedValue string, bankAccountID int) saqueFixtureResult {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/estabelecimentos/saques", session, map[string]any{
		"description":  description,
		"asked_value":  askedValue,
		"bank_account": bankAccountID,
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
