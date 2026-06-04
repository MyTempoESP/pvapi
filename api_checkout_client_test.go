package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

type chkCliFixtures struct {
	productID   int
	packageID   int
	shopSession apiSession
	admin       apiSession
}

func setupChkCliFixtures(t *testing.T, h *apiHarness, tag string, runID int64) chkCliFixtures {
	t.Helper()

	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja ChkCli %s %d", tag, runID)
	shopEmail := fmt.Sprintf("api-chkcli-%s-%d@example.com", tag, runID)
	h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	monitor := h.createMonitor(t, &shopSession,
		fmt.Sprintf("Monitor ChkCli %s %d", tag, runID),
		fmt.Sprintf("api-chkcli-mon-%s-%d@example.com", tag, runID),
		"27998209673")
	pickup := h.createPickup(t, &shopSession)

	cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-chkcli-%s-%d", tag, runID))
	productName := fmt.Sprintf("Produto ChkCli %s %d", tag, runID)
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

func TestCheckoutClientAutoCreation(t *testing.T) {
	h := newAPIHarness(t)
	baseRunID := time.Now().UnixNano()

	t.Run("happy_path_client_created", func(t *testing.T) {
		runID := baseRunID + 1
		fx := setupChkCliFixtures(t, h, "happy", runID)
		sessionID := fmt.Sprintf("chkcli-happy-%d", runID)

		clientEmail := fmt.Sprintf("chkcli-happy-%d@example.com", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-CHKCLI-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		assertStatus(t, rec, http.StatusOK)

		rec2 := h.doRequest(t, http.MethodGet, "/clientes", &fx.shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)
		body2 := decodeJSON(t, rec2)
		clients := asSlice(t, body2["clients"])
		var found bool
		for _, cl := range clients {
			client := asMap(t, cl)
			if asString(t, client["email"]) == clientEmail {
				found = true
				want := fmt.Sprintf("usuario-chkclihappy%d", runID)
				if asString(t, client["name"]) != want {
					t.Fatalf("expected fallback name %q, got %q", want, client["name"])
				}
				break
			}
		}
		if !found {
			t.Fatalf("client with email %q not found in client list", clientEmail)
		}
	})

	t.Run("existing_client_reused", func(t *testing.T) {
		runID := baseRunID + 2
		fx := setupChkCliFixtures(t, h, "reuse", runID)
		clientEmail := fmt.Sprintf("chkcli-reuse-%d@example.com", runID)

		// First checkout creates the client
		sessionID1 := fmt.Sprintf("chkcli-reuse-1-%d", runID)
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID1, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-REUSE1-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		assertStatus(t, rec1, http.StatusOK)

		// Second checkout reuses the same email — create a new product/consignment
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor Reuse2 %d", runID),
			fmt.Sprintf("api-chkcli-mon-reuse2-%d@example.com", runID),
			"27998209674")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-reuse2-%d", runID))
		prd2Name := fmt.Sprintf("Produto Reuse2 %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		sessionID2 := fmt.Sprintf("chkcli-reuse-2-%d", runID)
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID2, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-REUSE2-%d", runID),
			"items": []map[string]any{
				{"package_id": pkgID2, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		assertStatus(t, rec2, http.StatusOK)

		// Verify only one client exists for this email
		rec3 := h.doRequest(t, http.MethodGet, "/clientes", &fx.shopSession, "", nil)
		assertStatus(t, rec3, http.StatusOK)
		body3 := decodeJSON(t, rec3)
		clients := asSlice(t, body3["clients"])
		count := 0
		for _, cl := range clients {
			client := asMap(t, cl)
			if asString(t, client["email"]) == clientEmail {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("expected exactly 1 client with email %q, found %d", clientEmail, count)
		}
	})

	t.Run("missing_email_no_client_created", func(t *testing.T) {
		runID := baseRunID + 3
		fx := setupChkCliFixtures(t, h, "noemail", runID)
		sessionID := fmt.Sprintf("chkcli-noemail-%d", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-NOEMAIL-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if asString(t, body["status"]) != "paid" {
			t.Fatalf("expected status 'paid', got %v", body["status"])
		}
	})

	t.Run("client_name_provided", func(t *testing.T) {
		runID := baseRunID + 4
		fx := setupChkCliFixtures(t, h, "name", runID)
		sessionID := fmt.Sprintf("chkcli-name-%d", runID)

		clientEmail := fmt.Sprintf("chkcli-name-%d@example.com", runID)
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-NAME-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": clientEmail,
			"client_name":  "João Silva",
		})
		assertStatus(t, rec, http.StatusOK)

		rec2 := h.doRequest(t, http.MethodGet, "/clientes", &fx.shopSession, "", nil)
		assertStatus(t, rec2, http.StatusOK)
		body2 := decodeJSON(t, rec2)
		clients := asSlice(t, body2["clients"])
		var found bool
		for _, cl := range clients {
			client := asMap(t, cl)
			if asString(t, client["email"]) == clientEmail {
				found = true
				if asString(t, client["name"]) != "João Silva" {
					t.Fatalf("expected name 'João Silva', got %q", client["name"])
				}
				break
			}
		}
		if !found {
			t.Fatalf("client with email %q not found in client list", clientEmail)
		}
	})

	t.Run("client_creation_does_not_fail_checkout", func(t *testing.T) {
		runID := baseRunID + 5
		fx := setupChkCliFixtures(t, h, "failok", runID)
		sessionID := fmt.Sprintf("chkcli-failok-%d", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-FAILOK-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": "valid-but-non-existent@example.com",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if asString(t, body["status"]) != "paid" {
			t.Fatalf("expected status 'paid' even if client creation fails, got %v", body["status"])
		}
	})

	t.Run("pickup_confirmation_email_sent_on_new_client", func(t *testing.T) {
		t.Setenv("PV_API_LOGIN_LINK", "https://pvaauto.example.com/login")
		t.Setenv("PV_API_PASSWORD_RESET_LINK", "https://pvaauto.example.com/reset")

		runID := baseRunID + 6
		fx := setupChkCliFixtures(t, h, "email", runID)
		sessionID := fmt.Sprintf("chkcli-email-%d", runID)

		clientEmail := fmt.Sprintf("chkcli-email-%d@example.com", runID)
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-EMAIL-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": clientEmail,
			"client_name":  "Maria Oliveira",
		})
		assertStatus(t, rec, http.StatusOK)

		if h.mailer.SentConfirmation == nil {
			t.Fatalf("expected a pickup confirmation email to be sent")
		}

		email := h.mailer.SentConfirmation
		if email.OrderID == "" {
			t.Errorf("expected non-empty OrderID")
		}
		if email.QrCode == "" {
			t.Errorf("expected non-empty QrCode")
		}
		if !email.IsNewUser {
			t.Errorf("expected IsNewUser=true for new client")
		}
		if email.Name != "Maria Oliveira" {
			t.Errorf("expected Name='Maria Oliveira', got %q", email.Name)
		}
		if email.Email != clientEmail {
			t.Errorf("expected Email=%q, got %q", clientEmail, email.Email)
		}
		if email.Password == "" {
			t.Errorf("expected non-empty Password for new user")
		}
		if email.LoginURL != "https://pvaauto.example.com/login" {
			t.Errorf("expected LoginURL=https://pvaauto.example.com/login, got %q", email.LoginURL)
		}
		if email.ResetURL != "https://pvaauto.example.com/reset" {
			t.Errorf("expected ResetURL=https://pvaauto.example.com/reset, got %q", email.ResetURL)
		}

		// Should include product and pickup info
		if len(email.Groups) == 0 {
			t.Errorf("expected at least one pickup group")
		}
		hasProduct := false
		hasPickup := false
		for _, g := range email.Groups {
			if g.PickupName != "" {
				hasPickup = true
			}
			for _, item := range g.Items {
				if strings.Contains(item.ProductName, "Produto ChkCli") {
					hasProduct = true
				}
			}
		}
		if !hasProduct {
			t.Errorf("expected email groups to contain a product name")
		}
		if !hasPickup {
			t.Errorf("expected email groups to contain a pickup location")
		}
	})

	t.Run("no_login_details_for_existing_client", func(t *testing.T) {
		t.Setenv("PV_API_LOGIN_LINK", "https://pvaauto.example.com/login")
		t.Setenv("PV_API_PASSWORD_RESET_LINK", "https://pvaauto.example.com/reset")

		runID := baseRunID + 7
		fx := setupChkCliFixtures(t, h, "reuse2", runID)
		clientEmail := fmt.Sprintf("chkcli-reuse2-%d@example.com", runID)

		// First checkout creates the client
		sessionID1 := fmt.Sprintf("chkcli-reuse2-1-%d", runID)
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID1, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-REUSE2A-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		assertStatus(t, rec1, http.StatusOK)

		// Create a second product/consignment for the second checkout
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor Reuse2B %d", runID),
			fmt.Sprintf("api-chkcli-mon-reuse2b-%d@example.com", runID),
			"27998209675")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-reuse2b-%d", runID))
		prd2Name := fmt.Sprintf("Produto Reuse2B %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		// Second checkout with same email — existing user, no login details
		sessionID2 := fmt.Sprintf("chkcli-reuse2-2-%d", runID)
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID2, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-REUSE2B-%d", runID),
			"items": []map[string]any{
				{"package_id": pkgID2, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		assertStatus(t, rec2, http.StatusOK)

		// The second checkout sends a confirmation email with IsNewUser=false and no password
		if h.mailer.SentConfirmation == nil {
			t.Fatalf("expected a confirmation email to be sent")
		}

		email := h.mailer.SentConfirmation
		if email.IsNewUser {
			t.Errorf("expected IsNewUser=false for existing client")
		}
		if email.Password != "" {
			t.Errorf("expected empty Password for existing client, got %q", email.Password)
		}
		if email.LoginURL != "" {
			t.Errorf("expected empty LoginURL for existing client, got %q", email.LoginURL)
		}
		if email.ResetURL != "" {
			t.Errorf("expected empty ResetURL for existing client, got %q", email.ResetURL)
		}

		// Should still have order info
		if email.OrderID == "" {
			t.Errorf("expected non-empty OrderID")
		}
		if len(email.Groups) == 0 {
			t.Errorf("expected at least one pickup group")
		}
	})

	t.Run("no_email_without_client_email", func(t *testing.T) {
		// Ensure env vars are set so the mailer would send if called
		t.Setenv("PV_API_LOGIN_LINK", "https://pvaauto.example.com/login")
		t.Setenv("PV_API_PASSWORD_RESET_LINK", "https://pvaauto.example.com/reset")

		runID := baseRunID + 8
		fx := setupChkCliFixtures(t, h, "noemail2", runID)
		sessionID := fmt.Sprintf("chkcli-noemail2-%d", runID)

		prevData := h.mailer.SentConfirmation

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-NOEMAIL2-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		if h.mailer.SentConfirmation != prevData {
			t.Fatalf("expected no new emails sent when client_email is missing")
		}
	})
}
