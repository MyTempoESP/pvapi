package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestMonitor(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	admin := h.login(t, "sistemas@lzero.org", "TestingPassword")
	shopName := fmt.Sprintf("Loja Monitor %d", runID)
	shopEmail := fmt.Sprintf("api-monitor-shop-%d@example.com", runID)
	_ = h.createShop(t, &admin, shopName, shopEmail, "11988887777")
	shopSession := h.login(t, shopEmail, "11988887777")

	monitorName := fmt.Sprintf("Pedro %d", runID)
	monitorEmail := fmt.Sprintf("api-monitor-%d@example.com", runID)

	t.Run("create_monitor", func(t *testing.T) {
		result := h.createMonitor(t, &shopSession, monitorName, monitorEmail, "27998209673")
		if result.ID == 0 {
			t.Fatal("expected non-zero monitor ID")
		}
	})

	t.Run("list_monitors", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/monitores", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("delete_monitor", func(t *testing.T) {
		deleteEmail := fmt.Sprintf("api-del-monitor-%d@example.com", runID)
		deleteName := fmt.Sprintf("Para Deletar %d", runID)

		result := h.createMonitor(t, &shopSession, deleteName, deleteEmail, "11999999999")

		rec := h.doRequest(t, http.MethodDelete, fmt.Sprintf("/monitores/%d", result.ID), &shopSession, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("monitor_login", func(t *testing.T) {
		loginEmail := fmt.Sprintf("api-monitor-login-%d@example.com", runID)
		result := h.createMonitor(t, &shopSession, monitorName+"-login", loginEmail, "27998209673")
		session := h.login(t, result.Email, result.Phone)
		if len(session.cookies) == 0 {
			t.Fatal("expected auth cookies for monitor login")
		}
	})

	t.Run("consignee_list_remessas_happy_path", func(t *testing.T) {
		monitorEmail := fmt.Sprintf("api-monitor-remessa-happy-%d@example.com", runID)
		monitorResult := h.createMonitor(t, &shopSession, "Monitor Remessa", monitorEmail, "27998209673")

		pickup := h.createPickup(t, &shopSession)
		consignment := h.createConsignment(t, &shopSession, pickup.ID, monitorResult.ID)

		// Add a product to the consignment so totalProducts is non-zero
		cat := h.createCategory(t, &shopSession, "Cat Remessa Stats")
		prod := h.createProduct(t, &shopSession, "Prod Remessa Stats", cat.ID)
		h.approveProduct(t, &admin, prod.ID)
		h.addProductToConsignment(t, &shopSession, consignment.ID, prod.ID, prod.Name)

		monitorSession := h.login(t, monitorResult.Email, monitorResult.Phone)
		rec := h.doRequest(t, http.MethodGet, "/monitores/me/remessas", &monitorSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		data := asSlice(t, body["data"])
		if len(data) == 0 {
			t.Fatal("expected at least one remessa in the list")
		}
		item := findObjectByIntField(t, data, "id", consignment.ID)
		if item == nil {
			t.Fatalf("expected remessa %d in the list", consignment.ID)
		}

		// Assert new aggregate fields
		totalProducts := asInt(t, item["totalProducts"])
		soldProducts := asInt(t, item["soldProducts"])
		if totalProducts != 2 {
			t.Fatalf("expected totalProducts=2, got %d", totalProducts)
		}
		if soldProducts != 0 {
			t.Fatalf("expected soldProducts=0, got %d", soldProducts)
		}

		// Assert camelCase fields match frontend contract
		if _, ok := item["createdAt"]; !ok {
			t.Fatal("expected createdAt field")
		}
		if _, ok := item["destination"]; !ok {
			t.Fatal("expected destination field")
		}
	})

	t.Run("consignee_list_remessas_empty", func(t *testing.T) {
		monitorEmail := fmt.Sprintf("api-monitor-remessa-empty-%d@example.com", runID)
		monitorResult := h.createMonitor(t, &shopSession, "Monitor Empty", monitorEmail, "27998209673")

		monitorSession := h.login(t, monitorResult.Email, monitorResult.Phone)
		rec := h.doRequest(t, http.MethodGet, "/monitores/me/remessas", &monitorSession, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		data := asSlice(t, body["data"])
		if len(data) != 0 {
			t.Fatal("expected empty remessa list")
		}
	})

	t.Run("consignee_list_remessas_shop_rejected", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/monitores/me/remessas", &shopSession, "", nil)
		assertStatus(t, rec, http.StatusForbidden)
	})
}
