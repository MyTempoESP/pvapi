package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestCheckoutShopFreeze(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")

	// ── Shop 1: FIXED commission (createConsignment defaults) ──────────
	shop1Name := fmt.Sprintf("Loja ChkShop1 %d", runID)
	shop1Email := fmt.Sprintf("api-chkshop1-%d@example.com", runID)
	shop1 := h.createShop(t, &admin, shop1Name, shop1Email, "11988887771")
	shop1Session := h.login(t, shop1Email, "11988887771")

	monitor1 := h.createMonitor(t, &shop1Session,
		fmt.Sprintf("Monitor ChkShop1 %d", runID),
		fmt.Sprintf("api-chkshop1-mon-%d@example.com", runID),
		"27998209671")
	pickup1 := h.createPickup(t, &shop1Session)

	cat1 := h.createCategory(t, &shop1Session, fmt.Sprintf("cat-chkshop1-%d", runID))
	prd1Name := fmt.Sprintf("Produto ChkShop1 %d", runID)
	prd1 := h.createProduct(t, &shop1Session, prd1Name, cat1.ID)
	h.approveProduct(t, &admin, prd1.ID)

	cons1 := h.createConsignment(t, &shop1Session, pickup1.ID, monitor1.ID)
	h.addProductToConsignment(t, &shop1Session, cons1.ID, prd1.ID, prd1Name)
	pkgID1 := h.getPackageID(t, prd1.ID, runID)

	// ── Shop 2: PERCENT commission ─────────────────────────────────────
	shop2Name := fmt.Sprintf("Loja ChkShop2 %d", runID)
	shop2Email := fmt.Sprintf("api-chkshop2-%d@example.com", runID)
	shop2 := h.createShop(t, &admin, shop2Name, shop2Email, "11988887772")
	shop2Session := h.login(t, shop2Email, "11988887772")

	monitor2 := h.createMonitor(t, &shop2Session,
		fmt.Sprintf("Monitor ChkShop2 %d", runID),
		fmt.Sprintf("api-chkshop2-mon-%d@example.com", runID),
		"27998209672")
	pickup2 := h.createPickup(t, &shop2Session)

	cat2 := h.createCategory(t, &shop2Session, fmt.Sprintf("cat-chkshop2-%d", runID))
	prd2Name := fmt.Sprintf("Produto ChkShop2 %d", runID)
	prd2 := h.createProduct(t, &shop2Session, prd2Name, cat2.ID)
	h.approveProduct(t, &admin, prd2.ID)

	// Use PERCENT commission for shop2
	rec := h.doJSON(t, http.MethodPost, "/remessas", &shop2Session, map[string]any{
		"destinationId":           pickup2.ID,
		"monitorId":               monitor2.ID,
		"status":                  "CONFERIR - DEVOLUCAO",
		"monitorProfitPercentage": 3,
	})
	assertStatus(t, rec, http.StatusOK)
	body := decodeJSON(t, rec)
	cons2ID := asInt(t, body["id"])

	h.addProductToConsignment(t, &shop2Session, cons2ID, prd2.ID, prd2Name)

	// getPackageID uses the default consignment setup, so
	// for shop2 we need to get the package ID manually
	pkgSessionID := fmt.Sprintf("pkg-lookup-shop2-%d", runID)
	h.createCartItem(t, pkgSessionID, prd2.ID, 1)
	optsRec := h.doRequest(t, http.MethodGet,
		fmt.Sprintf("/carrinho/pacotes?session_id=%s", pkgSessionID),
		nil, "", nil)
	assertStatus(t, optsRec, http.StatusOK)
	optsBody := decodeJSON(t, optsRec)
	optsProducts := asSlice(t, optsBody["products"])
	if len(optsProducts) == 0 {
		t.Fatal("no products in package options for shop2")
	}
	shop2Pkg := findObjectByIntField(t, optsProducts, "product_id", prd2.ID)
	shop2Packages := asSlice(t, shop2Pkg["packages"])
	if len(shop2Packages) == 0 {
		t.Fatal("no packages for shop2")
	}
	pkgID2 := asInt(t, asMap(t, shop2Packages[0])["package_id"])

	// ── Verify consignment details ──────────────────────────────────────
	// The fixtures use:
	//   product price=2, package price=3, quantity=2 (addProductToConsignment)
	//   platform_cut=5 (percentage)
	//   Shop1 consignment: comission_value=3, comission_type='FIXED'
	//   Shop2 consignment: comission_value=3, comission_type='PERCENT'

	sessionID := fmt.Sprintf("chkshop-checkout-%d", runID)
	h.createCartItem(t, sessionID, prd1.ID, 1)
	h.createCartItem(t, sessionID, prd2.ID, 1)

	// Checkout with value_paid different from actual to test discrepancy
	externalRef := fmt.Sprintf("EXT-REF-CS-%d", runID)
	chkRec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
		"value_paid":         "99.99",
		"external_reference": externalRef,
		"items": []map[string]any{
			{"package_id": pkgID1, "quantity": 1},
			{"package_id": pkgID2, "quantity": 1},
		},
	})
	assertStatus(t, chkRec, http.StatusOK)
	chkBody := decodeJSON(t, chkRec)
	uniqueID := asString(t, chkBody["unique_id"])

	t.Run("checkout_shop_rows_created", func(t *testing.T) {
		rows, err := h.pool.Query(
			context.Background(),
			`SELECT shop_id, shop_value, shop_tax, shop_commission
			 FROM checkout_shop
			 WHERE checkout_unique_id = $1::uuid
			 ORDER BY shop_id`,
			uniqueID,
		)
		if err != nil {
			t.Fatalf("query checkout_shop: %v", err)
		}
		defer rows.Close()

		type csRow struct {
			ShopID         int     `db:"shop_id"`
			ShopValue      float64 `db:"shop_value"`
			ShopTax        float64 `db:"shop_tax"`
			ShopCommission float64 `db:"shop_commission"`
		}

		var shopRows []csRow
		for rows.Next() {
			var r csRow
			if err := rows.Scan(&r.ShopID, &r.ShopValue, &r.ShopTax, &r.ShopCommission); err != nil {
				t.Fatalf("scan checkout_shop row: %v", err)
			}
			shopRows = append(shopRows, r)
		}

		if len(shopRows) != 2 {
			t.Fatalf("expected 2 checkout_shop rows (one per shop), got %d", len(shopRows))
		}

		// Shop1: FIXED commission (comission_value=3)
		// package price=3, checkout quantity=1
		// shop_value = 3.00 * 1 = 3.00
		// shop_tax = (5/100) * 3.00 * 1 = 0.15
		// commission = 3.00 + (3.00 * 1) = 6.00
		if shopRows[0].ShopID == shop1.ID || shopRows[1].ShopID == shop1.ID {
			var s1 csRow
			if shopRows[0].ShopID == shop1.ID {
				s1 = shopRows[0]
			} else {
				s1 = shopRows[1]
			}
			if s1.ShopValue != 3.00 {
				t.Errorf("shop1: expected shop_value=3.00, got %v", s1.ShopValue)
			}
			if s1.ShopTax != 0.15 {
				t.Errorf("shop1: expected shop_tax=0.15, got %v", s1.ShopTax)
			}
			if s1.ShopCommission != 6.00 {
				t.Errorf("shop1: expected shop_commission=6.00, got %v", s1.ShopCommission)
			}
		}

		// Shop2: PERCENT commission (comission_value=3)
		// shop_value = 3.00 * 1 = 3.00
		// shop_tax = (5/100) * 3.00 * 1 = 0.15
		// commission = (3.00/100) * 3.00 * 1 = 0.09
		if shopRows[0].ShopID == shop2.ID || shopRows[1].ShopID == shop2.ID {
			var s2 csRow
			if shopRows[0].ShopID == shop2.ID {
				s2 = shopRows[0]
			} else {
				s2 = shopRows[1]
			}
			if s2.ShopValue != 3.00 {
				t.Errorf("shop2: expected shop_value=3.00, got %v", s2.ShopValue)
			}
			if s2.ShopTax != 0.15 {
				t.Errorf("shop2: expected shop_tax=0.15, got %v", s2.ShopTax)
			}
			if s2.ShopCommission != 0.09 {
				t.Errorf("shop2: expected shop_commission=0.09, got %v", s2.ShopCommission)
			}
		}
	})

	t.Run("value_discrepancy_computed", func(t *testing.T) {
		var discrepancy float64
		err := h.pool.QueryRow(
			context.Background(),
			`SELECT value_discrepancy FROM checkouts WHERE unique_id = $1::uuid`,
			uniqueID,
		).Scan(&discrepancy)
		if err != nil {
			t.Fatalf("query value_discrepancy: %v", err)
		}

		// value_paid = 99.99, shop_value shop1=3.00 + shop2=3.00 = 6.00
		// discrepancy = 99.99 - 6.00 = 93.99
		if discrepancy != 93.99 {
			t.Errorf("expected value_discrepancy=93.99, got %v", discrepancy)
		}
	})

	t.Run("shop_checkout_report_uses_shop_value", func(t *testing.T) {
		// Shop1 views its checkout report
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/checkouts", &shop1Session, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		checkouts := asSlice(t, body["checkouts"])

		// Should have 1 checkout in shop1's report
		if len(checkouts) != 1 {
			t.Fatalf("expected 1 checkout in shop1 report, got %d: %s", len(checkouts), rec.Body.String())
		}

		co := asMap(t, checkouts[0])
		val := asString(t, co["value"])
		if val != "3" && val != "3.00" {
			t.Errorf("shop1: expected checkout value=3.00 (shop_value), got %q", val)
		}

		tax := asString(t, co["tax"])
		if tax != "0.15" {
			t.Errorf("shop1: expected tax=0.15, got %q", tax)
		}

		comm := asString(t, co["commission"])
		if comm != "6" && comm != "6.00" {
			t.Errorf("shop1: expected commission=6.00, got %q", comm)
		}

		// The response should NOT contain value_paid or value_discrepancy fields
		if _, ok := co["value_paid"]; ok {
			t.Error("shop checkout report should not expose value_paid")
		}
		if _, ok := co["value_discrepancy"]; ok {
			t.Error("shop checkout report should not expose value_discrepancy")
		}
	})

	t.Run("totals_use_shop_value", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/estabelecimentos/checkouts", &shop1Session, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		totals := asMap(t, body["totals"])

		totalStr := asString(t, totals["total"])
		taxesStr := asString(t, totals["taxes"])
		commStr := asString(t, totals["commissions"])

		if totalStr != "3" && totalStr != "3.00" {
			t.Errorf("shop1 totals: expected total=3.00, got %q", totalStr)
		}
		if taxesStr != "0.15" {
			t.Errorf("shop1 totals: expected taxes=0.15, got %q", taxesStr)
		}
		if commStr != "6" && commStr != "6.00" {
			t.Errorf("shop1 totals: expected commissions=6.00, got %q", commStr)
		}
	})

	t.Run("available_balance_uses_checkout_shop", func(t *testing.T) {
		// Shop1 available balance = total - tax - commission - saques
		// = 3.00 - 0.15 - 6.00 - 0 = -3.15
		var balance float64
		// We need a transaction to use LockShopRowQuery — use h.pool and exec inline
		tx, err := h.pool.Begin(context.Background())
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()

		// Lock shop row
		_, err = tx.Exec(context.Background(),
			`SELECT id FROM shops WHERE id = $1 FOR UPDATE`, shop1.ID)
		if err != nil {
			t.Fatalf("lock shop: %v", err)
		}

		err = tx.QueryRow(context.Background(),
			`SELECT COALESCE(SUM(cs.shop_value - cs.shop_tax - cs.shop_commission), 0)
			 FROM checkout_shop cs
			 WHERE cs.shop_id = $1`,
			shop1.ID,
		).Scan(&balance)
		if err != nil {
			t.Fatalf("query available: %v", err)
		}

		if balance != -3.15 {
			t.Errorf("shop1: expected available balance=-3.15, got %v", balance)
		}
	})
}

func TestCheckoutShopClientID(t *testing.T) {
	h := newAPIHarness(t)
	baseRunID := time.Now().UnixNano()

	// Setup: one shop with one product in a consignment
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja ChkShopCli %d", baseRunID)
	shopEmail := fmt.Sprintf("api-chkshopcli-%d@example.com", baseRunID)
	h.createShop(t, &admin, shopName, shopEmail, "11988887790")
	shopSession := h.login(t, shopEmail, "11988887790")

	monitor := h.createMonitor(t, &shopSession,
		fmt.Sprintf("Monitor ChkShopCli %d", baseRunID),
		fmt.Sprintf("api-chkshopcli-mon-%d@example.com", baseRunID),
		"27998209690")
	pickup := h.createPickup(t, &shopSession)
	cat := h.createCategory(t, &shopSession, fmt.Sprintf("cat-chkshopcli-%d", baseRunID))
	prdName := fmt.Sprintf("Produto ChkShopCli %d", baseRunID)
	prd := h.createProduct(t, &shopSession, prdName, cat.ID)
	h.approveProduct(t, &admin, prd.ID)
	cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
	h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, prdName)

	t.Run("populates_client_id_when_email_provided", func(t *testing.T) {
		runID := baseRunID + 1
		pkgID := h.getPackageID(t, prd.ID, runID)
		sessionID := fmt.Sprintf("chkshopcli-email-%d", runID)
		h.createCartItem(t, sessionID, prd.ID, 1)

		clientEmail := fmt.Sprintf("chkshopcli-email-%d@example.com", runID)
		externalRef := fmt.Sprintf("EXT-REF-CLI-%d", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": externalRef,
			"items": []map[string]any{
				{"package_id": pkgID, "quantity": 1},
			},
			"client_email": clientEmail,
		})
		assertStatus(t, rec, http.StatusOK)

		// Query checkout_shop rows for this checkout and verify client_id
		type csClientRow struct {
			ClientID *int `db:"client_id"`
		}
		rows, err := h.pool.Query(
			context.Background(),
			`SELECT cs.client_id
			 FROM checkout_shop cs
			 JOIN checkouts chk ON chk.id = cs.checkout_id
			 WHERE chk.external_reference = $1`,
			externalRef,
		)
		if err != nil {
			t.Fatalf("query checkout_shop: %v", err)
		}
		defer rows.Close()

		var csRows []csClientRow
		for rows.Next() {
			var r csClientRow
			if err := rows.Scan(&r.ClientID); err != nil {
				t.Fatalf("scan checkout_shop row: %v", err)
			}
			csRows = append(csRows, r)
		}

		if len(csRows) == 0 {
			t.Fatal("expected at least one checkout_shop row")
		}
		for i, r := range csRows {
			if r.ClientID == nil {
				t.Errorf("checkout_shop[%d]: expected non-nil client_id, got NULL", i)
			}
		}
	})

	t.Run("leaves_client_id_null_when_no_email", func(t *testing.T) {
		runID := baseRunID + 2
		pkgID := h.getPackageID(t, prd.ID, runID)
		sessionID := fmt.Sprintf("chkshopcli-noemail-%d", runID)
		h.createCartItem(t, sessionID, prd.ID, 1)

		externalRef := fmt.Sprintf("EXT-REF-NOCLI-%d", runID)

		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm?session_id="+sessionID, nil, map[string]any{
			"value_paid":         "3.00",
			"external_reference": externalRef,
			"items": []map[string]any{
				{"package_id": pkgID, "quantity": 1},
			},
		})
		assertStatus(t, rec, http.StatusOK)

		type csClientRow struct {
			ClientID *int `db:"client_id"`
		}
		rows, err := h.pool.Query(
			context.Background(),
			`SELECT cs.client_id
			 FROM checkout_shop cs
			 JOIN checkouts chk ON chk.id = cs.checkout_id
			 WHERE chk.external_reference = $1`,
			externalRef,
		)
		if err != nil {
			t.Fatalf("query checkout_shop: %v", err)
		}
		defer rows.Close()

		var csRows []csClientRow
		for rows.Next() {
			var r csClientRow
			if err := rows.Scan(&r.ClientID); err != nil {
				t.Fatalf("scan checkout_shop row: %v", err)
			}
			csRows = append(csRows, r)
		}

		if len(csRows) == 0 {
			t.Fatal("expected at least one checkout_shop row")
		}
		for i, r := range csRows {
			if r.ClientID != nil {
				t.Errorf("checkout_shop[%d]: expected NULL client_id, got %d", i, *r.ClientID)
			}
		}
	})
}
