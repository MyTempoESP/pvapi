package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"linhazero.com/pvapi/models/checkoutmodel"
)

// setupClientOverviewFixtures creates a shop, product, consignment, and
// returns the package ID plus the admin and shop sessions for further use.
func setupClientOverviewFixtures(t *testing.T, h *apiHarness, tag string, runID int64) chkCliFixtures {
	t.Helper()
	return setupChkCliFixtures(t, h, tag, runID)
}

// registerClientUser creates a user via /register, inserts a client record
// via SQL, and logs in. Returns the client session and client ID.
func registerClientUser(t *testing.T, h *apiHarness, email, name, password string) (apiSession, int) {
	t.Helper()

	rec := h.doJSON(t, http.MethodPost, "/register", nil, map[string]any{
		"name":     name,
		"email":    email,
		"password": password,
	})
	assertStatus(t, rec, http.StatusOK)

	// Look up user ID to create client record
	var userID int
	err := h.pool.QueryRow(context.Background(), `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	if err != nil {
		t.Fatalf("lookup user %q: %v", email, err)
	}

	var clientID int
	err = h.pool.QueryRow(context.Background(),
		`INSERT INTO clients (user_id) VALUES ($1) RETURNING id`, userID,
	).Scan(&clientID)
	if err != nil {
		t.Fatalf("create client for user %d: %v", userID, err)
	}

	session := h.login(t, email, password)
	return session, clientID
}

func TestClientOverview(t *testing.T) {
	h := newAPIHarness(t)
	baseRunID := time.Now().UnixNano()

	t.Run("unauthenticated_returns_401", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/usuario/cliente/overview", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("non_client_user_returns_404", func(t *testing.T) {
		// Register a user without a client record
		email := fmt.Sprintf("overview-nonclient-%d@example.com", baseRunID)
		rec := h.doJSON(t, http.MethodPost, "/register", nil, map[string]any{
			"name":     "Non Client",
			"email":    email,
			"password": "TestingPassword",
		})
		assertStatus(t, rec, http.StatusOK)

		session := h.login(t, email, "TestingPassword")

		rec2 := h.doRequest(t, http.MethodGet, "/usuario/cliente/overview", &session, "", nil)
		assertStatus(t, rec2, http.StatusNotFound)
	})

	t.Run("client_with_no_checkouts", func(t *testing.T) {
		runID := baseRunID + 100
		email := fmt.Sprintf("overview-empty-%d@example.com", runID)
		clientSession, _ := registerClientUser(t, h, email, "Cliente Vazio", "TestingPassword")

		rec := h.doRequest(t, http.MethodGet, "/usuario/cliente/overview", &clientSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if asString(t, body["message"]) != "visão geral carregada" {
			t.Fatalf("expected message 'visão geral carregada', got %q", body["message"])
		}
		if asInt(t, body["orders"]) != 0 {
			t.Fatalf("expected orders=0, got %v", body["orders"])
		}
		if asInt(t, body["pending"]) != 0 {
			t.Fatalf("expected pending=0, got %v", body["pending"])
		}
		pkgs := asSlice(t, body["packages"])
		if len(pkgs) != 0 {
			t.Fatalf("expected empty packages, got %d items", len(pkgs))
		}
	})

	t.Run("client_with_one_checkout_multiple_packages", func(t *testing.T) {
		runID := baseRunID + 200
		fx := setupClientOverviewFixtures(t, h, "ov-multi", runID)
		clientEmail := fmt.Sprintf("overview-multi-%d@example.com", runID)
		clientSession, _ := registerClientUser(t, h, clientEmail, "Cliente Multi", "TestingPassword")

		sessionID := fmt.Sprintf("ov-multi-%d", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, h.defaultAPISession, map[string]any{
			"value_paid":         "6.00",
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 2},
			},
			"client_email": clientEmail,
			"client_name":  "Cliente Multi",
		})
		assertStatus(t, rec, http.StatusOK)

		rec2 := h.doRequest(t, http.MethodGet, "/usuario/cliente/overview", &clientSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)

		body := decodeJSON(t, rec2)
		if asInt(t, body["orders"]) != 1 {
			t.Fatalf("expected orders=1, got %v", body["orders"])
		}
		if asInt(t, body["pending"]) != 1 {
			t.Fatalf("expected pending=1, got %v", body["pending"])
		}

		pkgs := asSlice(t, body["packages"])
		if len(pkgs) != 1 {
			t.Fatalf("expected 1 package, got %d", len(pkgs))
		}

		pkg := asMap(t, pkgs[0])
		if _, ok := pkg["uuid"]; !ok {
			t.Errorf("expected package uuid field")
		}
		if _, ok := pkg["date"]; !ok {
			t.Errorf("expected package date field")
		}
		if asString(t, pkg["status"]) != "pending" {
			t.Errorf("expected status 'pending', got %q", pkg["status"])
		}
		if _, ok := pkg["value"]; !ok {
			t.Errorf("expected package value field")
		}
		if asInt(t, pkg["quantity"]) != 2 {
			t.Errorf("expected quantity 2, got %v", pkg["quantity"])
		}

		// Verify pickup structure
		pickup := asMap(t, pkg["pickup"])
		if _, ok := pickup["id"]; !ok {
			t.Errorf("expected pickup.id")
		}
		if _, ok := pickup["name"]; !ok {
			t.Errorf("expected pickup.name")
		}
		if _, ok := pickup["datetime"]; !ok {
			t.Errorf("expected pickup.datetime")
		}
		addr := asMap(t, pickup["address"])
		if _, ok := addr["street"]; !ok {
			t.Errorf("expected pickup.address.street")
		}

		// Verify product structure
		product := asMap(t, pkg["product"])
		if _, ok := product["id"]; !ok {
			t.Errorf("expected product.id")
		}
		if _, ok := product["name"]; !ok {
			t.Errorf("expected product.name")
		}
		if _, ok := product["code"]; !ok {
			t.Errorf("expected product.code")
		}
	})

	t.Run("client_mixed_pending_and_received", func(t *testing.T) {
		runID := baseRunID + 300
		fx := setupClientOverviewFixtures(t, h, "ov-mix", runID)
		clientEmail := fmt.Sprintf("overview-mix-%d@example.com", runID)
		clientSession, clientID := registerClientUser(t, h, clientEmail, "Cliente Mix", "TestingPassword")

		// First checkout — will be updated to 'recebido' after
		sessionID1 := fmt.Sprintf("ov-mix-1-%d", runID)
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID1, h.defaultAPISession, map[string]any{
			"value_paid":         "3.00",
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1},
			},
			"client_email": clientEmail,
			"client_name":  "Cliente Mix",
		})
		assertStatus(t, rec1, http.StatusOK)

		// Mark first package as received (direct SQL)
		_, err := h.pool.Exec(context.Background(),
			`UPDATE package_reserved SET status = 'recebido'
			 WHERE checkout_id = (
			   SELECT chk.id FROM checkouts chk
			   WHERE chk.client_id = $1
			   ORDER BY chk.processed_at ASC
			   LIMIT 1
			 )`, clientID)
		if err != nil {
			t.Fatalf("update package status to recebido: %v", err)
		}

		// Second checkout — stays 'pending'
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor OVMix2 %d", runID),
			fmt.Sprintf("api-ovmix-mon2-%d@example.com", runID),
			"27998209675")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-ovmix2-%d", runID))
		prd2Name := fmt.Sprintf("Produto OVMix2 %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		savedWindow := checkoutmodel.RateLimitWindowSeconds
		checkoutmodel.RateLimitWindowSeconds = 0

		sessionID2 := fmt.Sprintf("ov-mix-2-%d", runID)
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID2, h.defaultAPISession, map[string]any{
			"value_paid":         "3.00",
			"items": []map[string]any{
				{"id": pkgID2, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		checkoutmodel.RateLimitWindowSeconds = savedWindow
		assertStatus(t, rec2, http.StatusOK)

		rec3 := h.doRequest(t, http.MethodGet, "/usuario/cliente/overview", &clientSession, "", nil)
		assertStatus(t, rec3, http.StatusOK)

		body := decodeJSON(t, rec3)
		if asInt(t, body["orders"]) != 2 {
			t.Fatalf("expected orders=2, got %v", body["orders"])
		}
		if asInt(t, body["pending"]) != 1 {
			t.Fatalf("expected pending=1, got %v", body["pending"])
		}

		pkgs := asSlice(t, body["packages"])
		if len(pkgs) != 2 {
			t.Fatalf("expected 2 packages, got %d", len(pkgs))
		}

		// Most recent checkout should be first
		pkg1 := asMap(t, pkgs[0])
		if asString(t, pkg1["status"]) != "pending" {
			t.Errorf("expected first package status 'pending', got %q", pkg1["status"])
		}

		pkg2 := asMap(t, pkgs[1])
		if asString(t, pkg2["status"]) != "recebido" {
			t.Errorf("expected second package status 'recebido', got %q", pkg2["status"])
		}
	})
}
