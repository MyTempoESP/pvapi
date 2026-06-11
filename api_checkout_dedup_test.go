package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

type chkRateLimitFixtures struct {
	packageID int
	admin     apiSession
}

func setupChkRateLimitFixtures(t *testing.T, h *apiHarness, tag string, runID int64) chkRateLimitFixtures {
	t.Helper()

	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja ChkRL %s %d", tag, runID)
	shopEmail := fmt.Sprintf("api-chkrl-%s-%d@example.com", tag, runID)
	h.createShop(t, &admin, shopName, shopEmail, "11988887770")
	shopSession := h.login(t, shopEmail, "11988887770")

	monitor := h.createMonitor(t, &shopSession,
		fmt.Sprintf("Monitor ChkRL %s %d", tag, runID),
		fmt.Sprintf("api-chkrl-mon-%s-%d@example.com", tag, runID),
		"27998209670")
	pickup := h.createPickup(t, &shopSession)

	cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-chkrl-%s-%d", tag, runID))
	productName := fmt.Sprintf("Produto ChkRL %s %d", tag, runID)
	prd := h.createProduct(t, &shopSession, productName, cat.ID)
	h.approveProduct(t, &admin, prd.ID)

	cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
	h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, productName)

	packageID := h.getPackageID(t, prd.ID, runID)

	return chkRateLimitFixtures{
		packageID: packageID,
		admin:     admin,
	}
}

func TestCheckoutRateLimit(t *testing.T) {
	h := newAPIHarness(t)
	baseRunID := time.Now().UnixNano()

	t.Run("identical_payload_within_window_returns_429", func(t *testing.T) {
		runID := baseRunID + 1
		fx := setupChkRateLimitFixtures(t, h, "rldedup", runID)

		items := []map[string]any{{"package_id": fx.packageID, "quantity": 1}}
		extras := map[string]any{"client_email": fmt.Sprintf("rldedup-%d@example.com", runID)}

		// First request — succeeds
		rec1 := h.confirmCheckout(t, "3.00", items, extras)
		assertStatus(t, rec1, http.StatusOK)

		// Second request — same payload, same client, within 60s window → 429
		rec2 := h.confirmCheckout(t, "3.00", items, extras)
		assertStatus(t, rec2, http.StatusTooManyRequests)

		body2 := decodeJSON(t, rec2)
		success, ok := body2["success"].(bool)
		if ok && success {
			t.Fatal("expected success=false for rate-limited request")
		}
		if asInt(t, body2["retry_after"]) <= 0 || asInt(t, body2["retry_after"]) > 60 {
			t.Fatalf("expected retry_after between 1 and 60, got %d", asInt(t, body2["retry_after"]))
		}
		if asString(t, body2["message"]) == "" {
			t.Fatal("expected non-empty message for rate-limit error")
		}
	})

	t.Run("different_payload_same_client_within_window_returns_429", func(t *testing.T) {
		runID := baseRunID + 2
		fx := setupChkRateLimitFixtures(t, h, "rldiff", runID)

		email := fmt.Sprintf("rldiff-%d@example.com", runID)
		items := []map[string]any{{"package_id": fx.packageID, "quantity": 1}}
		extras := map[string]any{"client_email": email}

		// First request — succeeds
		rec1 := h.confirmCheckout(t, "3.00", items, extras)
		assertStatus(t, rec1, http.StatusOK)

		// Second request — different value_paid, same client, within window → 429
		rec2 := h.confirmCheckout(t, "5.00", items, extras)
		assertStatus(t, rec2, http.StatusTooManyRequests)
	})

	t.Run("same_payload_different_client_succeeds", func(t *testing.T) {
		runID := baseRunID + 3
		fx := setupChkRateLimitFixtures(t, h, "rldiffcli", runID)

		// First client
		email1 := fmt.Sprintf("rldiffcli-a-%d@example.com", runID)
		items := []map[string]any{{"package_id": fx.packageID, "quantity": 1}}

		rec1 := h.confirmCheckout(t, "3.00", items, map[string]any{"client_email": email1})
		assertStatus(t, rec1, http.StatusOK)

		// Second client — same payload but different email → succeeds
		email2 := fmt.Sprintf("rldiffcli-b-%d@example.com", runID)
		rec2 := h.confirmCheckout(t, "3.00", items, map[string]any{"client_email": email2})
		assertStatus(t, rec2, http.StatusOK)

		if asString(t, rec1.Body.String()) == asString(t, rec2.Body.String()) {
			t.Fatal("expected different responses for different clients")
		}
	})

	t.Run("no_identity_returns_400", func(t *testing.T) {
		runID := baseRunID + 4
		fx := setupChkRateLimitFixtures(t, h, "rlnoident", runID)

		items := []map[string]any{{"package_id": fx.packageID, "quantity": 1}}

		// No email, no CPF, no telephone
		rec := h.confirmCheckout(t, "3.00", items, map[string]any{})
		assertStatus(t, rec, http.StatusBadRequest)

		body := decodeJSON(t, rec)
		if asString(t, body["message"]) != "Informe email, CPF ou telefone para continuar" {
			t.Fatalf("expected identity validation message, got %q", asString(t, body["message"]))
		}
	})

	t.Run("cpf_only_same_cpf_within_window_returns_429", func(t *testing.T) {
		runID := baseRunID + 5
		fx := setupChkRateLimitFixtures(t, h, "rlcpf", runID)

		cpf := fmt.Sprintf("%d", 10000000000+runID)
		if len(cpf) > 11 {
			cpf = cpf[len(cpf)-11:]
		}

		items := []map[string]any{{"package_id": fx.packageID, "quantity": 1}}
		extras := map[string]any{"cpf": cpf}

		// First request — succeeds (client created via CPF)
		rec1 := h.confirmCheckout(t, "3.00", items, extras)
		assertStatus(t, rec1, http.StatusOK)

		// Second request — same CPF within window → 429
		rec2 := h.confirmCheckout(t, "3.00", items, extras)
		assertStatus(t, rec2, http.StatusTooManyRequests)
	})

	t.Run("telephone_only_same_phone_within_window_returns_429", func(t *testing.T) {
		runID := baseRunID + 6
		fx := setupChkRateLimitFixtures(t, h, "rltel", runID)

		telephone := fmt.Sprintf("119%08d", runID%100000000)

		items := []map[string]any{{"package_id": fx.packageID, "quantity": 1}}
		extras := map[string]any{"telephone": telephone}

		rec1 := h.confirmCheckout(t, "3.00", items, extras)
		assertStatus(t, rec1, http.StatusOK)

		rec2 := h.confirmCheckout(t, "3.00", items, extras)
		assertStatus(t, rec2, http.StatusTooManyRequests)
	})
}
