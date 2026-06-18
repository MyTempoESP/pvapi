package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"linhazero.com/pvapi/models/checkoutmodel"
)

func TestClientSales(t *testing.T) {
	h := newAPIHarness(t)
	baseRunID := time.Now().UnixNano()

	t.Run("unauthenticated_returns_401", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/clients/1/sales", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("non_shop_user_returns_403", func(t *testing.T) {
		// Register a user (not shop)
		email := fmt.Sprintf("sales-nonshop-%d@example.com", baseRunID)
		rec := h.doJSON(t, http.MethodPost, "/register", nil, map[string]any{
			"name":     "Non Shop User",
			"email":    email,
			"password": "TestingPassword",
		})
		assertStatus(t, rec, http.StatusOK)

		session := h.login(t, email, "TestingPassword")

		rec2 := h.doRequest(t, http.MethodGet, "/clients/1/sales", &session, "", nil)
		// Non-shop users are blocked at the authz layer (403)
		assertStatus(t, rec2, http.StatusForbidden)
	})

	t.Run("client_does_not_exist_returns_404", func(t *testing.T) {
		// Need a shop session first
		admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
		shopEmail := fmt.Sprintf("sales-nonexist-%d@example.com", baseRunID)
		h.createShop(t, &admin, fmt.Sprintf("Loja NonExist %d", baseRunID),
			shopEmail, "11988887777")
		shopSession := h.login(t, shopEmail, "11988887777")

		rec := h.doRequest(t, http.MethodGet, "/clients/999999/sales", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)
	})

	t.Run("client_with_no_sales_returns_empty_array", func(t *testing.T) {
		runID := baseRunID + 100
		admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
		shopEmail := fmt.Sprintf("sales-empty-%d@example.com", runID)
		h.createShop(t, &admin, fmt.Sprintf("Loja Empty %d", runID),
			shopEmail, "11988887777")
		shopSession := h.login(t, shopEmail, "11988887777")

		// Create a client user and client record (no sales)
		clientEmail := fmt.Sprintf("sales-empty-cli-%d@example.com", runID)
		rec := h.doJSON(t, http.MethodPost, "/register", nil, map[string]any{
			"name":     "Empty Client",
			"email":    clientEmail,
			"password": "TestingPassword",
		})
		assertStatus(t, rec, http.StatusOK)

		userID := lookupUserID(t, h, clientEmail)

		// Create a client record for the user (no checkouts)
		_, err := h.pool.Exec(context.Background(),
			`INSERT INTO clients (user_id) VALUES ($1)`, userID)
		if err != nil {
			t.Fatalf("create client: %v", err)
		}

		rec2 := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/clients/%d/sales", userID),
			&shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		// Response must be a bare JSON array
		var result []any
		if err := decodeRawArray(t, rec2, &result); err != nil {
			t.Fatalf("decode array: %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("expected empty array, got %d items", len(result))
		}
	})

	t.Run("client_with_one_checkout", func(t *testing.T) {
		runID := baseRunID + 200
		fx := setupClientOverviewFixtures(t, h, "sales-one", runID)
		clientEmail := fmt.Sprintf("sales-one-%d@example.com", runID)

		sessionID := fmt.Sprintf("sales-one-%d", runID)
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, h.defaultAPISession, map[string]any{
			"value_paid":         "3.00",
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"client_name":  "Maria Cliente",
		})
		assertStatus(t, rec, http.StatusOK)

		userID := lookupUserID(t, h, clientEmail)
		clientID := lookupClientID(t, h, clientEmail)

		rec2 := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/clients/%d/sales", userID),
			&fx.shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		var sales []map[string]any
		if err := decodeRawArray(t, rec2, &sales); err != nil {
			t.Fatalf("decode array: %v", err)
		}
		if len(sales) != 1 {
			t.Fatalf("expected 1 sale, got %d", len(sales))
		}

		sale := sales[0]
		if _, ok := sale["id"]; !ok {
			t.Fatalf("expected sale.id")
		}
		if asInt(t, sale["clientId"]) != clientID {
			t.Fatalf("expected clientId %d, got %v", clientID, sale["clientId"])
		}
		if _, ok := sale["total"]; !ok {
			t.Fatalf("expected sale.total")
		}
		if _, ok := sale["date"]; !ok {
			t.Fatalf("expected sale.date")
		}
		if _, ok := sale["status"]; !ok {
			t.Fatalf("expected sale.status")
		}
		products := asSlice(t, sale["products"])
		if len(products) == 0 {
			t.Fatalf("expected non-empty products array")
		}
	})

	t.Run("client_with_multiple_checkouts", func(t *testing.T) {
		runID := baseRunID + 300
		fx := setupClientOverviewFixtures(t, h, "sales-multi", runID)
		clientEmail := fmt.Sprintf("sales-multi-%d@example.com", runID)

		// First checkout
		sessionID1 := fmt.Sprintf("sales-multi-1-%d", runID)
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID1, h.defaultAPISession, map[string]any{
			"value_paid":         "3.00",
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"client_name":  "Joana Multi",
		})
		assertStatus(t, rec1, http.StatusOK)

		// Second checkout with a different product
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor SM2 %d", runID),
			fmt.Sprintf("api-sales-multi-mon2-%d@example.com", runID),
			"27998209676")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-sales-multi2-%d", runID))
		prd2Name := fmt.Sprintf("Produto SM2 %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		savedWindow := checkoutmodel.RateLimitWindowSeconds
		checkoutmodel.RateLimitWindowSeconds = 0

		sessionID2 := fmt.Sprintf("sales-multi-2-%d", runID)
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID2, h.defaultAPISession, map[string]any{
			"value_paid":         "3.00",
			"items": []map[string]any{
				{"id": pkgID2, "quantity": 1, "retirada": pickup2.ID},
			},
			"client_email": clientEmail,
		})
		checkoutmodel.RateLimitWindowSeconds = savedWindow
		assertStatus(t, rec2, http.StatusOK)

		userID := lookupUserID(t, h, clientEmail)

		rec3 := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/clients/%d/sales", userID),
			&fx.shopSession, "", nil)
		assertStatus(t, rec3, http.StatusOK)

		var sales []map[string]any
		if err := decodeRawArray(t, rec3, &sales); err != nil {
			t.Fatalf("decode array: %v", err)
		}
		if len(sales) != 2 {
			t.Fatalf("expected 2 sales, got %d", len(sales))
		}

		// Should be ordered by date descending (most recent first)
		date1 := asString(t, sales[0]["date"])
		date2 := asString(t, sales[1]["date"])
		if date1 < date2 {
			t.Fatalf("expected sales ordered by date desc, got %s before %s", date1, date2)
		}
	})

	t.Run("client_mixed_pending_and_received_status", func(t *testing.T) {
		runID := baseRunID + 400
		fx := setupClientOverviewFixtures(t, h, "sales-mix", runID)
		clientEmail := fmt.Sprintf("sales-mix-%d@example.com", runID)

		// Checkout
		sessionID := fmt.Sprintf("sales-mix-%d", runID)
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, h.defaultAPISession, map[string]any{
			"value_paid":         "3.00",
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"client_name":  "Cliente Mix",
		})
		assertStatus(t, rec, http.StatusOK)

		userID := lookupUserID(t, h, clientEmail)

		// Initially status should be "Pendente" (package just created)
		rec2 := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/clients/%d/sales", userID),
			&fx.shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		var sales []map[string]any
		if err := decodeRawArray(t, rec2, &sales); err != nil {
			t.Fatalf("decode array: %v", err)
		}
		if len(sales) != 1 {
			t.Fatalf("expected 1 sale, got %d", len(sales))
		}
		if asString(t, sales[0]["status"]) != "Pendente" {
			t.Fatalf("expected status 'Pendente', got %q", sales[0]["status"])
		}

		// Mark all packages as received
		clientID := lookupClientID(t, h, clientEmail)
		_, err := h.pool.Exec(context.Background(),
			`UPDATE package_reserved SET status = 'recebido'
			 WHERE checkout_id = (
			   SELECT checkout_id FROM checkout_shop WHERE client_id = $1 LIMIT 1
			 )`, clientID)
		if err != nil {
			t.Fatalf("update package status to recebido: %v", err)
		}

		// Now status should be "Concluída"
		rec3 := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/clients/%d/sales", userID),
			&fx.shopSession, "", nil)
		assertStatus(t, rec3, http.StatusOK)

		var sales2 []map[string]any
		if err := decodeRawArray(t, rec3, &sales2); err != nil {
			t.Fatalf("decode array: %v", err)
		}
		if len(sales2) != 1 {
			t.Fatalf("expected 1 sale, got %d", len(sales2))
		}
		if asString(t, sales2[0]["status"]) != "Concluída" {
			t.Fatalf("expected status 'Concluída', got %q", sales2[0]["status"])
		}
	})
}

// lookupUserID queries the users table for a user by email, returning the user ID.
func lookupUserID(t *testing.T, h *apiHarness, email string) int {
	t.Helper()

	var userID int
	err := h.pool.QueryRow(context.Background(),
		`SELECT id FROM users WHERE email = $1`, email,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("lookup user id for %q: %v", email, err)
	}
	return userID
}

// lookupClientID queries the clients table for a client associated with the
// given user email, returning the client ID.
func lookupClientID(t *testing.T, h *apiHarness, email string) int {
	t.Helper()

	var clientID int
	err := h.pool.QueryRow(context.Background(), `
		SELECT c.id FROM clients c
		JOIN users u ON u.id = c.user_id
		WHERE u.email = $1
	`, email).Scan(&clientID)
	if err != nil {
		t.Fatalf("lookup client id for %q: %v", email, err)
	}
	return clientID
}

// decodeRawArray decodes a raw JSON array from the response into dst.
func decodeRawArray[T any](t *testing.T, rec *httptest.ResponseRecorder, dst *[]T) error {
	t.Helper()
	return json.NewDecoder(rec.Body).Decode(dst)
}
