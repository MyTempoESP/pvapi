package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func setupClientFixtures(t *testing.T, h *apiHarness, tag string, runID int64) chkCliFixtures {
	t.Helper()

	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja ClientF %s %d", tag, runID)
	shopEmail := fmt.Sprintf("api-clientf-%s-%d@example.com", tag, runID)
	h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	monitor := h.createMonitor(t, &shopSession,
		fmt.Sprintf("Monitor ClientF %s %d", tag, runID),
		fmt.Sprintf("api-clientf-mon-%s-%d@example.com", tag, runID),
		"27998209673")
	pickup := h.createPickup(t, &shopSession)

	cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-clientf-%s-%d", tag, runID))
	productName := fmt.Sprintf("Produto ClientF %s %d", tag, runID)
	prd := h.createProduct(t, &shopSession, productName, cat.ID)
	h.approveProduct(t, &admin, prd.ID)

	cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
	h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, productName)

	packageID := h.getPackageID(t, prd.ID, runID)

	return chkCliFixtures{
		productID:   prd.ID,
		packageID:   packageID,
		shopSession: shopSession,
		admin:       admin,
	}
}

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

	t.Run("list_clients_with_aggregates", func(t *testing.T) {
		runID := time.Now().UnixNano()
		fx := setupClientFixtures(t, h, "agg", runID)
		clientEmail := fmt.Sprintf("clientf-agg-%d@example.com", runID)

		// First checkout: 1x item at 3.00 each
		sessionID1 := fmt.Sprintf("clientf-agg-1-%d", runID)
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID1, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-AGG1-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": clientEmail,
			"client_name":  "Maria Agregada",
		})
		assertStatus(t, rec1, http.StatusOK)

		// Second checkout: 1x item at 3.00 each — same email
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor Agg2 %d", runID),
			fmt.Sprintf("api-clientf-mon-agg2-%d@example.com", runID),
			"27998209674")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-agg2-%d", runID))
		prd2Name := fmt.Sprintf("Produto Agg2 %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		sessionID2 := fmt.Sprintf("clientf-agg-2-%d", runID)
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID2, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-AGG2-%d", runID),
			"items": []map[string]any{
				{"package_id": pkgID2, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		assertStatus(t, rec2, http.StatusOK)

		// Fetch client list and verify aggregates
		rec3 := h.doRequest(t, http.MethodGet, "/clientes", &fx.shopSession, "", nil)
		assertStatus(t, rec3, http.StatusOK)
		body3 := decodeJSON(t, rec3)
		clients := asSlice(t, body3["clients"])

		client := findObjectByStringField(t, clients, "email", clientEmail)
		if client == nil {
			t.Fatalf("client with email %q not found", clientEmail)
		}

		// total_purchases should be "6.00" (2x 3.00)
		totalPurchases := asString(t, client["total_purchases"])
		if totalPurchases != "6.00" && totalPurchases != "6" {
			t.Fatalf("expected total_purchases '6.00', got %q", totalPurchases)
		}

		// last_purchase should be present and non-nil
		lastPurchase, ok := client["last_purchase"]
		if !ok || lastPurchase == nil {
			t.Fatalf("expected last_purchase to be present and non-nil, got %v", lastPurchase)
		}

		// last_purchase should be an ISO 8601 string
		lpStr := asString(t, lastPurchase)
		if !strings.Contains(lpStr, "T") {
			t.Fatalf("expected last_purchase to be ISO 8601, got %q", lpStr)
		}
	})
}
