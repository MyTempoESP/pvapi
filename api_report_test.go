package main

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestReport(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Report %d", runID)
	shopEmail := fmt.Sprintf("api-report-shop-%d@example.com", runID)
	shop := h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	monitorName := fmt.Sprintf("Monitor RP %d", runID)
	monitorEmail := fmt.Sprintf("api-rp-monitor-%d@example.com", runID)
	monitor := h.createMonitor(t, &shopSession, monitorName, monitorEmail, "27998209673")
	consigneeSession := h.login(t, monitor.Email, monitor.Phone)

	pickup := h.createPickup(t, &shopSession)

	cat := h.createCategory(t, &shopSession, "lixo-eletronico")
	productName := fmt.Sprintf("Produto RP %d", runID)
	prd := h.createProduct(t, &shopSession, productName, cat.ID)
	h.approveProduct(t, &admin, prd.ID)

	cons := h.createConsignment(t, &shopSession, pickup.ID, monitor.ID)
	h.addProductToConsignment(t, &shopSession, cons.ID, prd.ID, productName)

	otherShopName := fmt.Sprintf("Outra Loja RP %d", runID)
	otherShopEmail := fmt.Sprintf("api-rp-other-shop-%d@example.com", runID)
	otherShop := h.createShop(t, &admin, otherShopName, otherShopEmail, "11988887777")
	otherShopSession := h.login(t, otherShopEmail, "11988887777")

	otherCat := h.createCategory(t, &otherShopSession, "lixo-eletronico")
	otherProductName := fmt.Sprintf("Outro Produto RP %d", runID)
	otherPrd := h.createProduct(t, &otherShopSession, otherProductName, otherCat.ID)
	h.approveProduct(t, &admin, otherPrd.ID)

	otherMonitorName := fmt.Sprintf("Outro Monitor RP %d", runID)
	otherMonitorEmail := fmt.Sprintf("api-rp-other-monitor-%d@example.com", runID)
	otherMonitor := h.createMonitor(t, &otherShopSession, otherMonitorName, otherMonitorEmail, "27998209673")
	otherPickup := h.createPickup(t, &otherShopSession)
	otherCons := h.createConsignment(t, &otherShopSession, otherPickup.ID, otherMonitor.ID)

	_ = shop
	_ = otherShop
	_ = otherMonitor
	_ = otherCons

	t.Run("reject_nonexistent_product", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         99999,
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "product",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("reject_nonexistent_consignment", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         99999,
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "consignment",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("reject_unknown_scope", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         prd.ID,
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "invalid",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("create_product_scope_report", func(t *testing.T) {
		result := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "a",
			Priority:    "baixa",
			Scope:       "product",
		})
		if result.ID == 0 {
			t.Fatal("expected non-zero report ID")
		}
	})

	t.Run("create_report_string_orderId", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &shopSession, map[string]any{
			"orderId":         strconv.Itoa(prd.ID),
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "a",
			"priority":        "baixa",
			"occurrenceScope": "product",
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("create_report_shipment_scope", func(t *testing.T) {
		result := h.createReport(t, &shopSession, reportFixture{
			OrderID:     cons.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "a",
			Priority:    "baixa",
			Scope:       "shipment",
		})
		if result.ID == 0 {
			t.Fatal("expected non-zero report ID")
		}
	})

	t.Run("view_report_as_shop", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "view as shop",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asInt(t, report["id"]) != rpt.ID {
			t.Fatalf("expected report id %d, got body %s", rpt.ID, rec.Body.String())
		}
		if asString(t, report["orderName"]) != productName {
			t.Fatalf("expected orderName %q, got %v", productName, report["orderName"])
		}
	})

	t.Run("view_report_as_monitor", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "view as monitor",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &consigneeSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asInt(t, report["id"]) != rpt.ID {
			t.Fatalf("expected report id %d, got body %s", rpt.ID, rec.Body.String())
		}
		if asString(t, report["orderName"]) != productName {
			t.Fatalf("expected orderName %q, got %v", productName, report["orderName"])
		}
	})

	t.Run("list_shop_reports_and_monitor_list", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "list test",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doRequest(t, http.MethodGet, "/ocorrencias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		reports := asSlice(t, body["data"])
		if !containsID(reports, rpt.ID) {
			t.Fatalf("report %d not found in shop report list: %s", rpt.ID, rec.Body.String())
		}
		if !containsName(reports, productName) {
			t.Fatalf("orderName %q not found in shop report list: %s", productName, rec.Body.String())
		}

		rec = h.doRequest(t, http.MethodGet, "/ocorrencias", &consigneeSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		reports = asSlice(t, body["data"])
		if !containsID(reports, rpt.ID) {
			t.Fatalf("report %d not found in monitor report list: %s", rpt.ID, rec.Body.String())
		}
		if !containsName(reports, productName) {
			t.Fatalf("orderName %q not found in monitor report list: %s", productName, rec.Body.String())
		}
	})

	t.Run("update_status", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "update status",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, map[string]any{
			"status": "em-andamento",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asString(t, report["status"]) != "em-andamento" {
			t.Fatalf("expected status 'em-andamento', got %v", report["status"])
		}
	})

	t.Run("update_priority", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "update priority",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, map[string]any{
			"priority": "alta",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asString(t, report["priority"]) != "alta" {
			t.Fatalf("expected priority 'alta', got %v", report["priority"])
		}
	})

	t.Run("update_description", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "original description",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, map[string]any{
			"description": "descricao atualizada",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asString(t, report["description"]) != "descricao atualizada" {
			t.Fatalf("expected description 'descricao atualizada', got %v", report["description"])
		}
	})

	t.Run("update_type", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "update type",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, map[string]any{
			"type": "avaria",
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asString(t, report["type"]) != "avaria" {
			t.Fatalf("expected type 'avaria', got %v", report["type"])
		}
	})

	t.Run("delete_report", func(t *testing.T) {
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "to delete",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("consignee_create_product_report", func(t *testing.T) {
		result := h.createReport(t, &consigneeSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "consignee product report",
			Priority:    "baixa",
			Scope:       "product",
		})
		if result.ID == 0 {
			t.Fatal("expected non-zero report ID")
		}
	})

	t.Run("consignee_create_consignment_report", func(t *testing.T) {
		result := h.createReport(t, &consigneeSession, reportFixture{
			OrderID:     cons.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "consignee consignment report",
			Priority:    "baixa",
			Scope:       "consignment",
		})
		if result.ID == 0 {
			t.Fatal("expected non-zero report ID")
		}
	})

	t.Run("consignee_reject_product_nonexistent", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &consigneeSession, map[string]any{
			"orderId":         99999,
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "nonexistent product",
			"priority":        "baixa",
			"occurrenceScope": "product",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("consignee_reject_product_cross_shop", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &consigneeSession, map[string]any{
			"orderId":         otherPrd.ID,
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "other shop product",
			"priority":        "baixa",
			"occurrenceScope": "product",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("consignee_reject_consignment_nonexistent", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &consigneeSession, map[string]any{
			"orderId":         99999,
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "nonexistent consignment",
			"priority":        "baixa",
			"occurrenceScope": "consignment",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("consignee_reject_consignment_cross_monitor", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/ocorrencias", &consigneeSession, map[string]any{
			"orderId":         otherCons.ID,
			"monitor":         monitorName,
			"moment":          "exposicao",
			"type":            "falta_peca",
			"status":          "aberto",
			"description":     "other monitor consignment",
			"priority":        "baixa",
			"occurrenceScope": "consignment",
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("consignee_report_appears_in_both_lists", func(t *testing.T) {
		rpt := h.createReport(t, &consigneeSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "consignee listing test",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doRequest(t, http.MethodGet, "/ocorrencias", &consigneeSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		reports := asSlice(t, body["data"])
		if !containsID(reports, rpt.ID) {
			t.Fatalf("report %d not found in consignee report list: %s", rpt.ID, rec.Body.String())
		}

		rec = h.doRequest(t, http.MethodGet, "/ocorrencias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		reports = asSlice(t, body["data"])
		if !containsID(reports, rpt.ID) {
			t.Fatalf("report %d not found in shop report list: %s", rpt.ID, rec.Body.String())
		}
	})

	t.Run("consignment_orderName_with_three_pickups", func(t *testing.T) {
		pickup2 := h.createPickup(t, &shopSession)
		pickup3 := h.createPickup(t, &shopSession)

		cons3 := h.createConsignmentWithPickups(t, &shopSession, []int{pickup.ID, pickup2.ID, pickup3.ID}, monitor.ID)
		h.addProductToConsignment(t, &shopSession, cons3.ID, prd.ID, productName)

		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     cons3.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "consignment orderName test",
			Priority:    "baixa",
			Scope:       "consignment",
		})

		// View the report and check orderName
		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		orderName := asString(t, report["orderName"])

		// Expect "Monitor RP {runID} para {pickup1}, {pickup2}, {pickup3}"
		if !containsMonitorAndPickups(orderName, monitorName, 3) {
			t.Fatalf("expected formatted orderName with monitor and 3 pickups, got %q", orderName)
		}

		// List and check orderName
		rec = h.doRequest(t, http.MethodGet, "/ocorrencias", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body = decodeJSON(t, rec)
		reports := asSlice(t, body["data"])
		found := false
		for _, item := range reports {
			r := asMap(t, item)
			if asInt(t, r["id"]) == rpt.ID {
				if asString(t, r["orderName"]) != orderName {
					t.Fatalf("list orderName %q does not match view orderName %q", asString(t, r["orderName"]), orderName)
				}
				found = true
			}
		}
		if !found {
			t.Fatal("report not found in shop list")
		}
	})

	t.Run("consignment_orderName_no_pickups_returns_na", func(t *testing.T) {
		cons0 := h.createConsignmentWithPickups(t, &shopSession, []int{}, monitor.ID)
		h.addProductToConsignment(t, &shopSession, cons0.ID, prd.ID, productName)

		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     cons0.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "consignment no pickups",
			Priority:    "baixa",
			Scope:       "consignment",
		})

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asString(t, report["orderName"]) != "N/A" {
			t.Fatalf("expected orderName %q, got %q", "N/A", report["orderName"])
		}
	})

	t.Run("product_scope_report_still_uses_product_name", func(t *testing.T) {
		// Regression test: product-scope reports must still show product name
		rpt := h.createReport(t, &shopSession, reportFixture{
			OrderID:     prd.ID,
			Monitor:     monitorName,
			Moment:      "exposicao",
			Type:        "falta_peca",
			Status:      "aberto",
			Description: "product scope regression",
			Priority:    "baixa",
			Scope:       "product",
		})

		rec := h.doRequest(t, http.MethodGet, fmt.Sprintf("/ocorrencias/%d", rpt.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
		body := decodeJSON(t, rec)
		report := asMap(t, body["report"])
		if asString(t, report["orderName"]) != productName {
			t.Fatalf("expected orderName %q for product-scope report, got %q", productName, report["orderName"])
		}
	})
}
