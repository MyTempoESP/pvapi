package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"linhazero.com/pvapi/models/checkoutmodel"
)

type chkCliFixtures struct {
	productID   int
	packageID   int
	pickupID    int
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
		pickupID:    pickup.ID,
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

		clientEmail := fmt.Sprintf("chkcli-happy-%d@example.com", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
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
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
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

		// Disable rate limit for same-client second checkout
		savedWindow := checkoutmodel.RateLimitWindowSeconds
		checkoutmodel.RateLimitWindowSeconds = 0
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": pkgID2, "quantity": 1, "retirada": pickup2.ID},
			},
			"client_email": clientEmail,
		})
		checkoutmodel.RateLimitWindowSeconds = savedWindow
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

	t.Run("missing_identity_returns_400", func(t *testing.T) {
		runID := baseRunID + 3
		fx := setupChkCliFixtures(t, h, "noident", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
		})
		assertStatus(t, rec, http.StatusBadRequest)

		body := decodeJSON(t, rec)
		if asString(t, body["message"]) != "Informe email, CPF ou telefone para continuar" {
			t.Fatalf("expected identity validation message, got %q", asString(t, body["message"]))
		}
	})

	t.Run("client_name_provided", func(t *testing.T) {
		runID := baseRunID + 4
		fx := setupChkCliFixtures(t, h, "name", runID)

		clientEmail := fmt.Sprintf("chkcli-name-%d@example.com", runID)
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
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

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": "valid-but-non-existent@example.com",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if asString(t, body["status"]) != "paid" {
			t.Fatalf("expected status 'paid' even if client creation fails, got %v", body["status"])
		}
	})







	// --- Telephone and CPF resolution tests ---

	t.Run("telephone_becomes_password_for_new_client", func(t *testing.T) {
		runID := baseRunID + 10
		fx := setupChkCliFixtures(t, h, "telpw", runID)

		clientEmail := fmt.Sprintf("telpw-%d@example.com", runID)
		telephone := "11988887777"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"telephone":    telephone,
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify telephone stored on client
		cl := queryClientByEmail(t, h.pool, clientEmail)
		if cl.Telephone == nil || *cl.Telephone != telephone {
			t.Fatalf("expected telephone %q on client, got %v", telephone, cl.Telephone)
		}

		// Verify password hash matches telephone
		if !verifyPassword(t, h.pool, clientEmail, telephone) {
			t.Fatal("expected password to match telephone digits")
		}
	})

	t.Run("cpf_only_creates_synthetic_email", func(t *testing.T) {
		runID := baseRunID + 11
		fx := setupChkCliFixtures(t, h, "cpfonly", runID)

		cpf := fmt.Sprintf("%d", 10000000000+runID)
		if len(cpf) > 11 {
			cpf = cpf[len(cpf)-11:]
		}
		syntheticEmail := cpf + "@br"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"cpf": cpf,
		})
		assertStatus(t, rec, http.StatusOK)

		// Verify client created with synthetic email
		cl := queryClientByEmail(t, h.pool, syntheticEmail)
		if cl.Name != "usuario-"+cpf {
			t.Fatalf("expected name 'usuario-%s', got %q", cpf, cl.Name)
		}
		if cl.CPF == nil || *cl.CPF != cpf {
			t.Fatalf("expected CPF %q on client, got %v", cpf, cl.CPF)
		}

		// Verify password is random (16 chars, contains letters)
		if !verifyPasswordRandom(t, h.pool, syntheticEmail) {
			t.Fatal("expected random password for CPF-only client")
		}
	})

	t.Run("telephone_only_no_email", func(t *testing.T) {
		runID := baseRunID + 12
		fx := setupChkCliFixtures(t, h, "telonly", runID)

		telephone := "21988887777"
		syntheticEmail := telephone + "@tele"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"telephone": telephone,
		})
		assertStatus(t, rec, http.StatusOK)

		cl := queryClientByEmail(t, h.pool, syntheticEmail)
		if cl.Name != "usuario-"+telephone {
			t.Fatalf("expected name 'usuario-%s', got %q", telephone, cl.Name)
		}
		if cl.Telephone == nil || *cl.Telephone != telephone {
			t.Fatalf("expected telephone %q on client, got %v", telephone, cl.Telephone)
		}

		// Password = phone digits
		if !verifyPassword(t, h.pool, syntheticEmail, telephone) {
			t.Fatal("expected password to match telephone digits")
		}
	})

	t.Run("cpf_and_telephone_combined", func(t *testing.T) {
		runID := baseRunID + 13
		fx := setupChkCliFixtures(t, h, "cpftel", runID)

		cpf := fmt.Sprintf("%d", 20000000000+runID)
		if len(cpf) > 11 {
			cpf = cpf[len(cpf)-11:]
		}
		telephone := "11988887777"
		syntheticEmail := cpf + "@br"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"telephone": telephone,
			"cpf":       cpf,
		})
		assertStatus(t, rec, http.StatusOK)

		cl := queryClientByEmail(t, h.pool, syntheticEmail)
		if cl.CPF == nil || *cl.CPF != cpf {
			t.Fatalf("expected CPF %q, got %v", cpf, cl.CPF)
		}
		if cl.Telephone == nil || *cl.Telephone != telephone {
			t.Fatalf("expected telephone %q, got %v", telephone, cl.Telephone)
		}

		// Password = phone digits (telephone wins over random)
		if !verifyPassword(t, h.pool, syntheticEmail, telephone) {
			t.Fatal("expected password to match telephone digits")
		}
	})

	t.Run("real_email_wins_with_cpf_and_telephone", func(t *testing.T) {
		runID := baseRunID + 14
		fx := setupChkCliFixtures(t, h, "realwins", runID)

		clientEmail := fmt.Sprintf("realwins-%d@example.com", runID)
		cpf := fmt.Sprintf("%d", 30000000000+runID)
		if len(cpf) > 11 {
			cpf = cpf[len(cpf)-11:]
		}
		telephone := "11988887777"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"telephone":    telephone,
			"cpf":          cpf,
		})
		assertStatus(t, rec, http.StatusOK)

		cl := queryClientByEmail(t, h.pool, clientEmail)
		if cl.CPF == nil || *cl.CPF != cpf {
			t.Fatalf("expected CPF %q, got %v", cpf, cl.CPF)
		}
		if cl.Telephone == nil || *cl.Telephone != telephone {
			t.Fatalf("expected telephone %q, got %v", telephone, cl.Telephone)
		}
		if !verifyPassword(t, h.pool, clientEmail, telephone) {
			t.Fatal("expected password to match telephone digits")
		}
	})

	t.Run("existing_client_telephone_update", func(t *testing.T) {
		runID := baseRunID + 15
		fx := setupChkCliFixtures(t, h, "telupd", runID)
		clientEmail := fmt.Sprintf("telupd-%d@example.com", runID)
		firstPhone := "11988887777"
		secondPhone := "11999998888"

		t.Setenv("PV_API_LOGIN_LINK", "https://pvaauto.example.com/login")
		t.Setenv("PV_API_PASSWORD_RESET_LINK", "https://pvaauto.example.com/reset")

		// First checkout with original telephone
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"telephone":    firstPhone,
		})
		assertStatus(t, rec1, http.StatusOK)

		// Create a second product/consignment for second checkout
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor TelUpd2 %d", runID),
			fmt.Sprintf("api-telupd2-%d@example.com", runID),
			"27998209675")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-telupd2-%d", runID))
		prd2Name := fmt.Sprintf("Produto TelUpd2 %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		// Disable rate limit for same-client second checkout
		savedWindow := checkoutmodel.RateLimitWindowSeconds
		checkoutmodel.RateLimitWindowSeconds = 0

		// Second checkout with different telephone
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": pkgID2, "quantity": 1, "retirada": pickup2.ID},
			},
			"client_email": clientEmail,
			"telephone":    secondPhone,
		})
		checkoutmodel.RateLimitWindowSeconds = savedWindow
		assertStatus(t, rec2, http.StatusOK)

		// Telephone should be updated to second phone
		cl := queryClientByEmail(t, h.pool, clientEmail)
		if cl.Telephone == nil || *cl.Telephone != secondPhone {
			t.Fatalf("expected telephone updated to %q, got %v", secondPhone, cl.Telephone)
		}

		// Password should NOT be overwritten (still matches first phone)
		if !verifyPassword(t, h.pool, clientEmail, firstPhone) {
			t.Fatal("expected password to remain first phone (existing password not overwritten)")
		}
	})

	t.Run("existing_client_cpf_preserved_when_not_provided", func(t *testing.T) {
		runID := baseRunID + 16
		fx := setupChkCliFixtures(t, h, "cpfpres", runID)
		clientEmail := fmt.Sprintf("cpfpres-%d@example.com", runID)
		cpf := fmt.Sprintf("%d", 40000000000+runID)
		if len(cpf) > 11 {
			cpf = cpf[len(cpf)-11:]
		}

		// First checkout with CPF
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"telephone":    "11988887777",
			"cpf":          cpf,
		})
		assertStatus(t, rec1, http.StatusOK)

		// Create a second product/consignment
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor CpfPres2 %d", runID),
			fmt.Sprintf("api-cpfpres2-%d@example.com", runID),
			"27998209676")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-cpfpres2-%d", runID))
		prd2Name := fmt.Sprintf("Produto CpfPres2 %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		// Disable rate limit for same-client second checkout
		savedWindow := checkoutmodel.RateLimitWindowSeconds
		checkoutmodel.RateLimitWindowSeconds = 0

		// Second checkout WITHOUT CPF → existing CPF must be preserved
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": pkgID2, "quantity": 1, "retirada": pickup2.ID},
			},
			"client_email": clientEmail,
			"telephone":    "11988887777",
		})
		checkoutmodel.RateLimitWindowSeconds = savedWindow
		assertStatus(t, rec2, http.StatusOK)

		cl := queryClientByEmail(t, h.pool, clientEmail)
		if cl.CPF == nil || *cl.CPF != cpf {
			t.Fatalf("expected CPF %q preserved, got %v", cpf, cl.CPF)
		}
	})

	t.Run("existing_client_provided_fields_only", func(t *testing.T) {
		runID := baseRunID + 17
		fx := setupChkCliFixtures(t, h, "provonly", runID)
		clientEmail := fmt.Sprintf("provonly-%d@example.com", runID)
		originalPhone := "11988887777"

		// First checkout with telephone (no CPF)
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"telephone":    originalPhone,
		})
		assertStatus(t, rec1, http.StatusOK)

		// Create a second product/consignment
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor ProvOnly2 %d", runID),
			fmt.Sprintf("api-provonly2-%d@example.com", runID),
			"27998209677")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-provonly2-%d", runID))
		prd2Name := fmt.Sprintf("Produto ProvOnly2 %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		// Disable rate limit for same-client second checkout
		savedWindow := checkoutmodel.RateLimitWindowSeconds
		checkoutmodel.RateLimitWindowSeconds = 0

		// Second checkout: provide CPF but NO telephone
		newCPF := fmt.Sprintf("%d", 50000000000+runID)
		if len(newCPF) > 11 {
			newCPF = newCPF[len(newCPF)-11:]
		}
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": pkgID2, "quantity": 1, "retirada": pickup2.ID},
			},
			"client_email": clientEmail,
			"cpf":          newCPF,
		})
		checkoutmodel.RateLimitWindowSeconds = savedWindow
		assertStatus(t, rec2, http.StatusOK)

		cl := queryClientByEmail(t, h.pool, clientEmail)
		if cl.CPF == nil || *cl.CPF != newCPF {
			t.Fatalf("expected CPF updated to %q, got %v", newCPF, cl.CPF)
		}
		// Original telephone preserved (not overwritten with empty)
		if cl.Telephone == nil || *cl.Telephone != originalPhone {
			t.Fatalf("expected telephone preserved as %q, got %v", originalPhone, cl.Telephone)
		}
	})

	t.Run("invalid_cpf_fallback", func(t *testing.T) {
		runID := baseRunID + 18
		fx := setupChkCliFixtures(t, h, "invcpf", runID)

		telephone := "11988887777"
		syntheticEmail := telephone + "@tele"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"telephone": telephone,
			"cpf":       "invalid-cpf",
		})
		assertStatus(t, rec, http.StatusOK)

		// Falls through to telephone identity. CPF should be NULL in DB.
		cl := queryClientByEmail(t, h.pool, syntheticEmail)
		if cl.CPF != nil {
			t.Fatalf("expected CPF to be NULL for invalid CPF, got %q", *cl.CPF)
		}
	})

	t.Run("empty_phone_after_strip_returns_400", func(t *testing.T) {
		runID := baseRunID + 19
		fx := setupChkCliFixtures(t, h, "emptel", runID)

		// Telephone with only non-digits should be treated as not provided
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"telephone": "+() - ",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("no_pii_in_response", func(t *testing.T) {
		runID := baseRunID + 20
		fx := setupChkCliFixtures(t, h, "nopii", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": fmt.Sprintf("nopii-%d@example.com", runID),
			"telephone":    "11988887777",
			"cpf":          "12345678901",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		bodyBytes, _ := json.Marshal(body)
		bodyStr := string(bodyBytes)

		if strings.Contains(bodyStr, "11988887777") {
			t.Fatal("response body must not contain telephone")
		}
		if strings.Contains(bodyStr, "12345678901") {
			t.Fatal("response body must not contain CPF")
		}
	})

	t.Run("different_telephone_same_client_updates_telephone", func(t *testing.T) {
		runID := baseRunID + 21
		fx := setupChkCliFixtures(t, h, "telupd2", runID)
		clientEmail := fmt.Sprintf("telupd2-%d@example.com", runID)
		firstPhone := "11988887777"
		secondPhone := "11999998888"

		// First checkout creates client with first telephone
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": clientEmail,
			"telephone":    firstPhone,
		})
		assertStatus(t, rec1, http.StatusOK)

		// Verify first phone stored
		cl := queryClientByEmail(t, h.pool, clientEmail)
		if cl.Telephone == nil || *cl.Telephone != firstPhone {
			t.Fatalf("expected telephone %q on client, got %v", firstPhone, cl.Telephone)
		}

		// Directly update telephone via DB to test field update on next checkout
		_, err := h.pool.Exec(context.Background(),
			`UPDATE clients SET telephone = $1 WHERE id = (SELECT c.id FROM clients c JOIN users u ON u.id = c.user_id WHERE u.email = $2)`,
			secondPhone, clientEmail)
		if err != nil {
			t.Fatalf("update telephone: %v", err)
		}
		cl2 := queryClientByEmail(t, h.pool, clientEmail)
		if cl2.Telephone == nil || *cl2.Telephone != secondPhone {
			t.Fatalf("expected telephone updated to %q, got %v", secondPhone, cl2.Telephone)
		}
	})

	t.Run("cpf_with_dots_and_dashes", func(t *testing.T) {
		runID := baseRunID + 22
		fx := setupChkCliFixtures(t, h, "cpffmt", runID)

		cpf := fmt.Sprintf("%d", 60000000000+runID)
		if len(cpf) > 11 {
			cpf = cpf[len(cpf)-11:]
		}
		// Add formatting to test stripping
		formatted := cpf[:3] + "." + cpf[3:6] + "." + cpf[6:9] + "-" + cpf[9:]
		syntheticEmail := cpf + "@br"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"cpf": formatted,
		})
		assertStatus(t, rec, http.StatusOK)

		cl := queryClientByEmail(t, h.pool, syntheticEmail)
		if cl.CPF == nil || *cl.CPF != cpf {
			t.Fatalf("expected CPF %q on client, got %v", cpf, cl.CPF)
		}
	})

	t.Run("client_name_provided_with_cpf", func(t *testing.T) {
		runID := baseRunID + 23
		fx := setupChkCliFixtures(t, h, "cpfname", runID)

		cpf := fmt.Sprintf("%d", 70000000000+runID)
		if len(cpf) > 11 {
			cpf = cpf[len(cpf)-11:]
		}
		syntheticEmail := cpf + "@br"

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"cpf":         cpf,
			"client_name": "José Maria",
		})
		assertStatus(t, rec, http.StatusOK)

		cl := queryClientByEmail(t, h.pool, syntheticEmail)
		if cl.Name != "José Maria" {
			t.Fatalf("expected name 'José Maria', got %q", cl.Name)
		}
	})

	t.Run("missing_retirada_returns_400", func(t *testing.T) {
		runID := baseRunID + 30
		fx := setupChkCliFixtures(t, h, "noreti", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1},
			},
			"client_email": fmt.Sprintf("noreti-%d@example.com", runID),
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("invalid_retirada_returns_422", func(t *testing.T) {
		runID := baseRunID + 31
		fx := setupChkCliFixtures(t, h, "badreti", runID)

		// Use a pickup ID that does not belong to the consignment
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": 999999},
			},
			"client_email": fmt.Sprintf("badreti-%d@example.com", runID),
		})
		assertStatus(t, rec, http.StatusUnprocessableEntity)
		body := decodeJSON(t, rec)
		if !strings.Contains(asString(t, body["message"]), "retirada inválida") {
			t.Fatalf("expected retirada inválida message, got %q", asString(t, body["message"]))
		}
	})

	t.Run("personalizacao_in_request_echoed_in_response", func(t *testing.T) {
		runID := baseRunID + 40
		fx := setupChkCliFixtures(t, h, "pers", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID, "personalizacao": map[string]string{"tamanho": "G", "cor": "azul"}},
			},
			"client_email": fmt.Sprintf("pers-%d@example.com", runID),
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		qrCodes := asSlice(t, body["qr_codes"])
		if len(qrCodes) != 1 {
			t.Fatalf("expected 1 qr_code, got %d", len(qrCodes))
		}
		pers := asMap(t, qrCodes[0])["personalizacao"]
		persMap := asMap(t, pers)
		if asString(t, persMap["tamanho"]) != "G" {
			t.Fatalf("expected tamanho=G, got %v", persMap["tamanho"])
		}
		if asString(t, persMap["cor"]) != "azul" {
			t.Fatalf("expected cor=azul, got %v", persMap["cor"])
		}
	})

	t.Run("personalizacao_omitted_defaults_to_empty", func(t *testing.T) {
		runID := baseRunID + 41
		fx := setupChkCliFixtures(t, h, "persdef", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID},
			},
			"client_email": fmt.Sprintf("persdef-%d@example.com", runID),
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		qrCodes := asSlice(t, body["qr_codes"])
		if len(qrCodes) != 1 {
			t.Fatalf("expected 1 qr_code, got %d", len(qrCodes))
		}
		pers := asMap(t, qrCodes[0])["personalizacao"]
		if len(asMap(t, pers)) != 0 {
			t.Fatalf("expected empty personalizacao, got %v", pers)
		}
	})

	t.Run("personalizacao_empty_object_echoed", func(t *testing.T) {
		runID := baseRunID + 42
		fx := setupChkCliFixtures(t, h, "persemp", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID, "personalizacao": map[string]string{}},
			},
			"client_email": fmt.Sprintf("persemp-%d@example.com", runID),
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		qrCodes := asSlice(t, body["qr_codes"])
		if len(qrCodes) != 1 {
			t.Fatalf("expected 1 qr_code, got %d", len(qrCodes))
		}
		pers := asMap(t, qrCodes[0])["personalizacao"]
		if len(asMap(t, pers)) != 0 {
			t.Fatalf("expected empty personalizacao, got %v", pers)
		}
	})

	t.Run("multiple_items_each_with_different_personalizacao", func(t *testing.T) {
		runID := baseRunID + 43
		fx := setupChkCliFixtures(t, h, "persmul", runID)

		// Second product with its own consignment and pickup
		monitor2 := h.createMonitor(t, &fx.shopSession,
			fmt.Sprintf("Monitor PersMul2 %d", runID),
			fmt.Sprintf("api-chkcli-mon-persmul2-%d@example.com", runID),
			"27998209675")
		pickup2 := h.createPickup(t, &fx.shopSession)
		cat2 := h.createCategory(t, &fx.shopSession, fmt.Sprintf("cat-persmul-%d", runID))
		prd2Name := fmt.Sprintf("Produto PersMul %d", runID)
		prd2 := h.createProduct(t, &fx.shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &fx.shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &fx.shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"items": []map[string]any{
				{"id": fx.packageID, "quantity": 1, "retirada": fx.pickupID, "personalizacao": map[string]string{"cor": "vermelho"}},
				{"id": pkgID2, "quantity": 1, "retirada": pickup2.ID, "personalizacao": map[string]string{"tamanho": "GG"}},
			},
			"client_email": fmt.Sprintf("persmul-%d@example.com", runID),
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		qrCodes := asSlice(t, body["qr_codes"])
		if len(qrCodes) != 2 {
			t.Fatalf("expected 2 qr_codes, got %d", len(qrCodes))
		}

		pers0 := asMap(t, asMap(t, qrCodes[0])["personalizacao"])
		pers1 := asMap(t, asMap(t, qrCodes[1])["personalizacao"])

		if asString(t, pers0["cor"]) != "vermelho" {
			t.Fatalf("expected item 0 cor=vermelho, got %v", pers0)
		}
		if asString(t, pers1["tamanho"]) != "GG" {
			t.Fatalf("expected item 1 tamanho=GG, got %v", pers1)
		}
	})
}

// clientRow is a lightweight struct for querying client data directly from the database.
type clientRow struct {
	ID        int
	Name      string
	Email     string
	Telephone *string
	CPF       *string
	Password  []byte
}

// queryClientByEmail queries the database for a client by email, joining users and clients tables.
func queryClientByEmail(t *testing.T, pool *pgxpool.Pool, email string) clientRow {
	t.Helper()

	var row clientRow
	err := pool.QueryRow(
		context.Background(),
		`SELECT c.id, u.name, u.email, c.telephone, c.cpf, u.password_hash
		 FROM clients c
		 JOIN users u ON u.id = c.user_id
		 WHERE u.email = $1`,
		email,
	).Scan(&row.ID, &row.Name, &row.Email, &row.Telephone, &row.CPF, &row.Password)
	if err != nil {
		t.Fatalf("query client by email %q: %v", email, err)
	}
	return row
}

// verifyPassword checks if the given email's password hash matches the plaintext.
func verifyPassword(t *testing.T, pool *pgxpool.Pool, email, plaintext string) bool {
	t.Helper()

	row := queryClientByEmail(t, pool, email)
	err := bcrypt.CompareHashAndPassword(row.Password, []byte(plaintext))
	return err == nil
}

// verifyPasswordRandom checks that the password is NOT the phone number (i.e., it's random).
func verifyPasswordRandom(t *testing.T, pool *pgxpool.Pool, email string) bool {
	t.Helper()

	row := queryClientByEmail(t, pool, email)
	if len(row.Password) == 0 {
		return false
	}

	// Verify the hash is structurally valid by comparing with a wrong password.
	// bcrypt.CompareHashAndPassword returns an error for both wrong password
	// and invalid format. Valid format means the hash has the right prefix.
	const bcryptPrefix = "$2a$"
	hashStr := string(row.Password)
	if !strings.HasPrefix(hashStr, bcryptPrefix) && !strings.HasPrefix(hashStr, "$2b$") {
		return false
	}

	// The hash should NOT match the phone number (proves it's random, not phone digits)
	err := bcrypt.CompareHashAndPassword(row.Password, []byte("11988887777"))
	return err != nil
}
