package main

import (
	"fmt"
	"net/http"
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

		h.createCartItem(t, sessionID, fx.productID, 1)
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
		h.createCartItem(t, sessionID1, fx.productID, 1)
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
		h.createCartItem(t, sessionID2, prd2.ID, 1)
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
		h.createCartItem(t, sessionID, fx.productID, 1)

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
		h.createCartItem(t, sessionID, fx.productID, 1)

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
		h.createCartItem(t, sessionID, fx.productID, 1)

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

	t.Run("welcome_email_sent_on_client_creation", func(t *testing.T) {
		t.Setenv("PV_API_LOGIN_LINK", "https://pvaauto.example.com/login")
		t.Setenv("PV_API_PASSWORD_RESET_LINK", "https://pvaauto.example.com/reset")

		runID := baseRunID + 6
		fx := setupChkCliFixtures(t, h, "email", runID)
		sessionID := fmt.Sprintf("chkcli-email-%d", runID)
		h.createCartItem(t, sessionID, fx.productID, 1)

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

		if len(h.mailer.SentEmails) != 1 {
			t.Fatalf("expected exactly 1 welcome email sent, got %d", len(h.mailer.SentEmails))
		}

		email := h.mailer.SentEmails[0]
		if email.To != clientEmail {
			t.Errorf("expected email To=%q, got %q", clientEmail, email.To)
		}
		if email.Name != "Maria Oliveira" {
			t.Errorf("expected email Name='Maria Oliveira', got %q", email.Name)
		}
		if email.LoginURL != "https://pvaauto.example.com/login" {
			t.Errorf("expected LoginURL=https://pvaauto.example.com/login, got %q", email.LoginURL)
		}
		if email.ResetURL != "https://pvaauto.example.com/reset" {
			t.Errorf("expected ResetURL=https://pvaauto.example.com/reset, got %q", email.ResetURL)
		}
	})

	t.Run("no_welcome_email_without_client_email", func(t *testing.T) {
		// Ensure env vars are set so the mailer would send if called
		t.Setenv("PV_API_LOGIN_LINK", "https://pvaauto.example.com/login")
		t.Setenv("PV_API_PASSWORD_RESET_LINK", "https://pvaauto.example.com/reset")

		runID := baseRunID + 7
		fx := setupChkCliFixtures(t, h, "noemail2", runID)
		sessionID := fmt.Sprintf("chkcli-noemail2-%d", runID)
		h.createCartItem(t, sessionID, fx.productID, 1)

		prevCount := len(h.mailer.SentEmails)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-REF-NOEMAIL2-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		if len(h.mailer.SentEmails) != prevCount {
			t.Fatalf("expected no new emails sent, got %d new", len(h.mailer.SentEmails)-prevCount)
		}
	})
}
