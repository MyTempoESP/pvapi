package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

type chkDedupFixtures struct {
	packageID int
	admin     apiSession
}

func setupChkDedupFixtures(t *testing.T, h *apiHarness, tag string, runID int64) chkDedupFixtures {
	t.Helper()

	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja ChkDedup %s %d", tag, runID)
	shopEmail := fmt.Sprintf("api-chkdedup-%s-%d@example.com", tag, runID)
	h.createShop(t, &admin, shopName, shopEmail, "11988887770")
	shopSession := h.login(t, shopEmail, "11988887770")

	monitor := h.createMonitor(t, &shopSession,
		fmt.Sprintf("Monitor ChkDedup %s %d", tag, runID),
		fmt.Sprintf("api-chkdedup-mon-%s-%d@example.com", tag, runID),
		"27998209670")
	pickup := h.createPickup(t, &shopSession)

	cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-chkdedup-%s-%d", tag, runID))
	productName := fmt.Sprintf("Produto ChkDedup %s %d", tag, runID)
	prd := h.createProduct(t, &shopSession, productName, cat.ID)
	h.approveProduct(t, &admin, prd.ID)

	cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
	h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, productName)

	packageID := h.getPackageID(t, prd.ID, runID)

	return chkDedupFixtures{
		packageID: packageID,
		admin:     admin,
	}
}

func TestCheckoutContentHashDedup(t *testing.T) {
	h := newAPIHarness(t)
	baseRunID := time.Now().UnixNano()

	t.Run("identical_payload_returns_same_checkout", func(t *testing.T) {
		runID := baseRunID + 1
		fx := setupChkDedupFixtures(t, h, "dedup1", runID)

		body := map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-DEDUP1A-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": fmt.Sprintf("dedup1-%d@example.com", runID),
		}

		// First request
		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, body)
		assertStatus(t, rec1, http.StatusOK)
		r1 := decodeJSON(t, rec1)

		// Second request — same payload (same value_paid, items, email)
		body2 := map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-DEDUP1B-%d", runID), // different ref, same content
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": fmt.Sprintf("dedup1-%d@example.com", runID),
		}
		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, body2)
		assertStatus(t, rec2, http.StatusOK)
		r2 := decodeJSON(t, rec2)

		// Same unique_id and qr_code — idempotent
		if asString(t, r1["unique_id"]) != asString(t, r2["unique_id"]) {
			t.Fatalf("expected same unique_id for identical payload, got %q vs %q",
				r1["unique_id"], r2["unique_id"])
		}
		if asString(t, r1["qr_code"]) != asString(t, r2["qr_code"]) {
			t.Fatalf("expected same qr_code for identical payload")
		}
		if asString(t, r1["status"]) != "paid" || asString(t, r2["status"]) != "paid" {
			t.Fatalf("expected status 'paid' on both requests")
		}
	})

	t.Run("different_value_paid_produces_different_checkout", func(t *testing.T) {
		runID := baseRunID + 2
		fx := setupChkDedupFixtures(t, h, "dedup2", runID)

		email := fmt.Sprintf("dedup2-%d@example.com", runID)

		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-DEDUP2A-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": email,
		})
		assertStatus(t, rec1, http.StatusOK)
		r1 := decodeJSON(t, rec1)

		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "5.00",
			"external_reference": fmt.Sprintf("EXT-DEDUP2B-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": email,
		})
		assertStatus(t, rec2, http.StatusOK)
		r2 := decodeJSON(t, rec2)

		if asString(t, r1["unique_id"]) == asString(t, r2["unique_id"]) {
			t.Fatal("expected different unique_id for different value_paid")
		}
	})

	t.Run("different_items_produces_different_checkout", func(t *testing.T) {
		runID := baseRunID + 3
		fx := setupChkDedupFixtures(t, h, "dedup3", runID)

		// Create a second product/package for multi-item test
		shopSession := h.login(t,
			fmt.Sprintf("api-chkdedup-dedup3-%d@example.com", runID),
			"11988887770")
		monitor2 := h.createMonitor(t, &shopSession,
			fmt.Sprintf("Monitor Dedup3B %d", runID),
			fmt.Sprintf("api-chkdedup-mon2-%d@example.com", runID),
			"27998209671")
		pickup2 := h.createPickup(t, &shopSession)
		cat2 := h.createCategory(t, &shopSession, fmt.Sprintf("cat-dedup3b-%d", runID))
		prd2Name := fmt.Sprintf("Produto Dedup3B %d", runID)
		prd2 := h.createProduct(t, &shopSession, prd2Name, cat2.ID)
		h.approveProduct(t, &fx.admin, prd2.ID)
		cons2 := h.createConsignment(t, &shopSession, pickup2.ID, monitor2.ID)
		h.addProductToConsignment(t, &shopSession, cons2.ID, prd2.ID, prd2Name)
		pkgID2 := h.getPackageID(t, prd2.ID, runID)

		email := fmt.Sprintf("dedup3-%d@example.com", runID)

		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-DEDUP3A-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": email,
		})
		assertStatus(t, rec1, http.StatusOK)
		r1 := decodeJSON(t, rec1)

		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": fmt.Sprintf("EXT-DEDUP3B-%d", runID),
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
				{"package_id": pkgID2, "quantity": 1},
			},
			"client_email": email,
		})
		assertStatus(t, rec2, http.StatusOK)
		r2 := decodeJSON(t, rec2)

		if asString(t, r1["unique_id"]) == asString(t, r2["unique_id"]) {
			t.Fatal("expected different unique_id for different items")
		}
	})

	t.Run("same_external_reference_different_payloads_are_distinct", func(t *testing.T) {
		runID := baseRunID + 4
		fx := setupChkDedupFixtures(t, h, "dedup4", runID)

		sharedRef := fmt.Sprintf("EXT-SHARED-%d", runID)
		email1 := fmt.Sprintf("dedup4a-%d@example.com", runID)
		email2 := fmt.Sprintf("dedup4b-%d@example.com", runID)

		rec1 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": sharedRef,
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": email1,
		})
		assertStatus(t, rec1, http.StatusOK)
		r1 := decodeJSON(t, rec1)

		rec2 := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "5.00",
			"external_reference": sharedRef, // same ref, different payload
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
			"client_email": email2,
		})
		assertStatus(t, rec2, http.StatusOK)
		r2 := decodeJSON(t, rec2)

		if asString(t, r1["unique_id"]) == asString(t, r2["unique_id"]) {
			t.Fatal("expected different unique_id when same external_reference has different content")
		}
	})

	t.Run("missing_external_reference_succeeds", func(t *testing.T) {
		runID := baseRunID + 5
		fx := setupChkDedupFixtures(t, h, "noextref", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid": "3.00",
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		if asString(t, body["status"]) != "paid" {
			t.Fatalf("expected status 'paid', got %v", body["status"])
		}
		if asString(t, body["unique_id"]) == "" {
			t.Fatal("expected non-empty unique_id")
		}
	})

	t.Run("empty_external_reference_succeeds", func(t *testing.T) {
		runID := baseRunID + 6
		fx := setupChkDedupFixtures(t, h, "emptyextref", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": "",
			"items": []map[string]any{
				{"package_id": fx.packageID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		if asString(t, body["status"]) != "paid" {
			t.Fatalf("expected status 'paid', got %v", body["status"])
		}
		if asString(t, body["unique_id"]) == "" {
			t.Fatal("expected non-empty unique_id")
		}
	})
}
