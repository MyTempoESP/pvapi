package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCheckoutCancel(t *testing.T) {
	h := newAPIHarness(t)
	baseRunID := time.Now().UnixNano()

	type cancelFixtures struct {
		packageID    int
		packageID2   int
		uniqueID     string
		shopSession  apiSession
		admin        apiSession
		checkoutID   int
		pendingPkgID int
	}

	setup := func(t *testing.T, tag string, runID int64) cancelFixtures {
		t.Helper()

		admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
		shopName := fmt.Sprintf("Loja Cancel %s %d", tag, runID)
		shopEmail := fmt.Sprintf("api-cancel-%s-%d@example.com", tag, runID)
		h.createShop(t, &admin, shopName, shopEmail, "11988887760")
		shopSession := h.login(t, shopEmail, "11988887760")

		monitor := h.createMonitor(t, &shopSession,
			fmt.Sprintf("Monitor Cancel %s %d", tag, runID),
			fmt.Sprintf("api-cancel-mon-%s-%d@example.com", tag, runID),
			"27998209660")
		pickup := h.createPickup(t, &shopSession)

		cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-cancel-%s-%d", tag, runID))
		prdName := fmt.Sprintf("Produto Cancel %s %d", tag, runID)
		prd := h.createProduct(t, &shopSession, prdName, cat.ID)
		h.approveProduct(t, &admin, prd.ID)

		cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)

		// Add two different products to same consignment (different package rows)
		h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, prdName)
		pkgID := h.getPackageID(t, prd.ID, runID)

		prd2Name := fmt.Sprintf("Produto Cancel2 %s %d", tag, runID)
		prd2 := h.createProduct(t, &shopSession, prd2Name, cat.ID)
		h.approveProduct(t, &admin, prd2.ID)
		h.addProductToConsignment(t, &shopSession, cons.ID, prd2.ID, prd2Name)

		// Get pkgID2 from product detail
		var pkgID2 int
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/produtos/%d", prd2.ID), nil, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		product := asMap(t, body["data"])
		pickups := asSlice(t, product["pickups"])
		for _, p := range pickups {
			pk := asMap(t, p)
			if asInt(t, pk["package_id"]) != pkgID {
				pkgID2 = asInt(t, pk["package_id"])
				break
			}
		}
		if pkgID2 == 0 {
			t.Fatal("could not find second package ID")
		}

		chkRec := h.confirmCheckout(t, []map[string]any{
			{"id": pkgID, "quantity": 2, "retirada": pickup.ID},
			{"id": pkgID2, "quantity": 1, "retirada": pickup.ID},
		}, map[string]any{"client_email": fmt.Sprintf("cancel-%s-%d@example.com", tag, runID)})
		assertStatus(t, chkRec, http.StatusOK)
		chkBody := decodeJSON(t, chkRec)
		uniqueID := asString(t, chkBody["unique_id"])

		var checkoutID int
		err := h.pool.QueryRow(
			context.Background(),
			`SELECT id FROM checkouts WHERE unique_id = $1::uuid`,
			uniqueID,
		).Scan(&checkoutID)
		if err != nil {
			t.Fatalf("get checkout ID: %v", err)
		}

		// Get the first pending package_reserved ID
		var pendingPkgID int
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT id FROM package_reserved WHERE checkout_id = $1 AND status = 'pending' LIMIT 1`,
			checkoutID,
		).Scan(&pendingPkgID)
		if err != nil {
			t.Fatalf("get pending package_reserved ID: %v", err)
		}

		return cancelFixtures{
			packageID:    pkgID,
			packageID2:   pkgID2,
			uniqueID:     uniqueID,
			shopSession:  shopSession,
			admin:        admin,
			checkoutID:   checkoutID,
			pendingPkgID: pendingPkgID,
		}
	}

	t.Run("full_cancellation_deletes_checkout", func(t *testing.T) {
		runID := baseRunID + 1
		fx := setup(t, "full", runID)

		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/checkouts/%s", fx.uniqueID),
			h.defaultAPISession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		if asInt(t, body["cancelled_count"]) != 2 {
			t.Fatalf("expected cancelled_count=2, got %v", body["cancelled_count"])
		}
		if asInt(t, body["remaining_count"]) != 0 {
			t.Fatalf("expected remaining_count=0, got %v", body["remaining_count"])
		}

		// Checkout deleted
		var exists bool
		err := h.pool.QueryRow(
			context.Background(),
			`SELECT EXISTS(SELECT 1 FROM checkouts WHERE id = $1)`,
			fx.checkoutID,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("query checkout: %v", err)
		}
		if exists {
			t.Fatal("expected checkout to be deleted")
		}

		// Stock returned: package started at qty=2, checkout used 2, now restored to 2
		for _, pkgID := range []int{fx.packageID, fx.packageID2} {
			var qty int
			err := h.pool.QueryRow(
				context.Background(),
				`SELECT quantity FROM packages WHERE id = $1`, pkgID,
			).Scan(&qty)
			if err != nil {
				t.Fatalf("query package %d: %v", pkgID, err)
			}
			if qty != 2 {
				t.Fatalf("package %d: expected qty=2 (restored), got %d", pkgID, qty)
			}
		}

		// checkout_shop rows deleted
		var csCount int
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT COUNT(*) FROM checkout_shop WHERE checkout_id = $1`,
			fx.checkoutID,
		).Scan(&csCount)
		if err != nil {
			t.Fatalf("query checkout_shop: %v", err)
		}
		if csCount != 0 {
			t.Fatalf("expected 0 checkout_shop rows, got %d", csCount)
		}
	})

	t.Run("partial_cancellation_recomputes_checkout_shop", func(t *testing.T) {
		runID := baseRunID + 2
		fx := setup(t, "part", runID)

		// Mark one package_reserved as RECEIVED
		_, err := h.pool.Exec(
			context.Background(),
			`UPDATE package_reserved SET status = 'recebido' WHERE id = $1`,
			fx.pendingPkgID,
		)
		if err != nil {
			t.Fatalf("mark as received: %v", err)
		}

		// Find the OTHER package's ID (the one that stays pending and will be cancelled)
		var cancelledPkgID int
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT pr.package_id FROM package_reserved pr
			 WHERE pr.checkout_id = $1 AND pr.id != $2`,
			fx.checkoutID, fx.pendingPkgID,
		).Scan(&cancelledPkgID)
		if err != nil {
			t.Fatalf("find other package: %v", err)
		}

		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/checkouts/%s", fx.uniqueID),
			h.defaultAPISession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)

		// 2 pending cancelled, 1 received remains
		if asInt(t, body["cancelled_count"]) != 1 {
			t.Fatalf("expected cancelled_count=1, got %v", body["cancelled_count"])
		}
		if asInt(t, body["remaining_count"]) != 1 {
			t.Fatalf("expected remaining_count=1, got %v", body["remaining_count"])
		}

		var exists bool
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT EXISTS(SELECT 1 FROM checkouts WHERE id = $1)`,
			fx.checkoutID,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("query checkout: %v", err)
		}
		if !exists {
			t.Fatal("expected checkout to remain after partial cancel")
		}

		var csCount int
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT COUNT(*) FROM checkout_shop WHERE checkout_id = $1`,
			fx.checkoutID,
		).Scan(&csCount)
		if err != nil {
			t.Fatalf("query checkout_shop: %v", err)
		}
		if csCount != 1 {
			t.Fatalf("expected 1 checkout_shop row, got %d", csCount)
		}

		// The package that was NOT marked received was cancelled
		// (stock returned to 2).
		var qty int
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT quantity FROM packages WHERE id = $1`, cancelledPkgID,
		).Scan(&qty)
		if err != nil {
			t.Fatalf("query package: %v", err)
		}
		if qty != 2 {
			t.Fatalf("expected cancelled package qty=2 (restored), got %d for pkg %d", qty, cancelledPkgID)
		}
	})

	t.Run("noop_cancellation_all_received", func(t *testing.T) {
		runID := baseRunID + 3
		fx := setup(t, "noop", runID)

		// Mark all package_reserved as RECEIVED
		_, err := h.pool.Exec(
			context.Background(),
			`UPDATE package_reserved SET status = 'recebido' WHERE checkout_id = $1`,
			fx.checkoutID,
		)
		if err != nil {
			t.Fatalf("mark all as received: %v", err)
		}

		rec := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/checkouts/%s", fx.uniqueID),
			h.defaultAPISession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)

		if asInt(t, body["cancelled_count"]) != 0 {
			t.Fatalf("expected cancelled_count=0, got %v", body["cancelled_count"])
		}
		if asInt(t, body["remaining_count"]) != 2 {
			t.Fatalf("expected remaining_count=2, got %v", body["remaining_count"])
		}

		// Checkout preserved
		var exists bool
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT EXISTS(SELECT 1 FROM checkouts WHERE id = $1)`,
			fx.checkoutID,
		).Scan(&exists)
		if err != nil {
			t.Fatalf("query checkout: %v", err)
		}
		if !exists {
			t.Fatal("expected checkout to remain after noop cancel")
		}

		// Stock not returned (no packages were pending)
		var qty int
		err = h.pool.QueryRow(
			context.Background(),
			`SELECT quantity FROM packages WHERE id = $1`, fx.packageID,
		).Scan(&qty)
		if err != nil {
			t.Fatalf("query package: %v", err)
		}
		if qty != 0 {
			t.Fatalf("expected qty=0 (stock not restored), got %d", qty)
		}
	})

	t.Run("unknown_unique_id_returns_404", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodDelete,
			"/checkouts/00000000-0000-0000-0000-000000000000",
			h.defaultAPISession, "", nil)
		assertStatus(t, rec, http.StatusNotFound)

		body := decodeJSON(t, rec)
		if asBool(t, body["success"]) {
			t.Fatal("expected success=false for unknown checkout")
		}
	})

	t.Run("mismatched_api_client_returns_403", func(t *testing.T) {
		runID := baseRunID + 5
		fx := setup(t, "auth", runID)

		// Create a second API client via admin
		rec := h.doJSON(t, http.MethodPost, "/admin/api-clients", &fx.admin, map[string]any{
			"name":  "Other Cancel Client",
			"email": fmt.Sprintf("other-cancel-%d@test.com", runID),
		})
		assertStatus(t, rec, http.StatusOK)
		otherBody := decodeJSON(t, rec)
		otherToken := asString(t, otherBody["token"])
		otherSession := &apiSession{bearerToken: otherToken}

		rec2 := h.doRequest(t, http.MethodDelete,
			fmt.Sprintf("/checkouts/%s", fx.uniqueID),
			otherSession, "", nil)
		assertStatus(t, rec2, http.StatusForbidden)

		body := decodeJSON(t, rec2)
		if asBool(t, body["success"]) {
			t.Fatal("expected success=false for unauthorized cancel")
		}
	})
}
