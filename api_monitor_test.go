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
		confirmedProducts := asInt(t, item["confirmedProducts"])
		if totalProducts != 0 {
			t.Fatalf("expected totalProducts=0 (no sales), got %d", totalProducts)
		}
		if confirmedProducts != 0 {
			t.Fatalf("expected confirmedProducts=0, got %d", confirmedProducts)
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

	t.Run("update_monitor_name", func(t *testing.T) {
		updateEmail := fmt.Sprintf("api-upd-monitor-%d@example.com", runID)
		updateName := "Nome Original"
		result := h.createMonitor(t, &shopSession, updateName, updateEmail, "27998209673")

		newName := "Nome Atualizado"
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/monitores/%d", result.ID), &shopSession, map[string]any{
			"name": newName,
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		monitor := asMap(t, body["monitor"])
		gotName := asString(t, monitor["name"])
		if gotName != newName {
			t.Fatalf("expected name=%q, got %q", newName, gotName)
		}
		// Ensure other fields unchanged
		gotEmail := asString(t, monitor["email"])
		if gotEmail != updateEmail {
			t.Fatalf("expected email=%q, got %q", updateEmail, gotEmail)
		}
	})

	t.Run("update_monitor_email", func(t *testing.T) {
		updateEmail := fmt.Sprintf("api-upd-email-%d@example.com", runID)
		result := h.createMonitor(t, &shopSession, "Email Test", updateEmail, "27998209673")

		newEmail := fmt.Sprintf("api-upd-email-new-%d@example.com", runID)
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/monitores/%d", result.ID), &shopSession, map[string]any{
			"email": newEmail,
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		monitor := asMap(t, body["monitor"])
		gotEmail := asString(t, monitor["email"])
		if gotEmail != newEmail {
			t.Fatalf("expected email=%q, got %q", newEmail, gotEmail)
		}
	})

	t.Run("update_monitor_phone", func(t *testing.T) {
		result := h.createMonitor(t, &shopSession, "Phone Test", fmt.Sprintf("api-upd-phone-%d@example.com", runID), "27998209673")

		newPhone := "11999999999"
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/monitores/%d", result.ID), &shopSession, map[string]any{
			"phone": newPhone,
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		monitor := asMap(t, body["monitor"])
		gotPhone := asString(t, monitor["phone"])
		if gotPhone != newPhone {
			t.Fatalf("expected phone=%q, got %q", newPhone, gotPhone)
		}
	})

	t.Run("update_monitor_address", func(t *testing.T) {
		result := h.createMonitor(t, &shopSession, "Addr Test", fmt.Sprintf("api-upd-addr-%d@example.com", runID), "27998209673")

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/monitores/%d", result.ID), &shopSession, map[string]any{
			"address": map[string]any{
				"street":       "Rua Nova Atualizada",
				"number":       "123",
				"neighborhood": "Centro",
				"city":         "São Paulo",
				"state":        "SP",
				"cep":          "01001000",
			},
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		monitor := asMap(t, body["monitor"])
		addr := asMap(t, monitor["address"])
		gotStreet := asString(t, addr["street"])
		if gotStreet != "Rua Nova Atualizada" {
			t.Fatalf("expected street=Rua Nova Atualizada, got %q", gotStreet)
		}
		gotNumber := asString(t, addr["number"])
		if gotNumber != "123" {
			t.Fatalf("expected number=123, got %q", gotNumber)
		}
	})

	t.Run("update_monitor_empty_body", func(t *testing.T) {
		result := h.createMonitor(t, &shopSession, "Empty Body", fmt.Sprintf("api-upd-empty-%d@example.com", runID), "27998209673")

		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/monitores/%d", result.ID), &shopSession, map[string]any{})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		monitor := asMap(t, body["monitor"])
		gotName := asString(t, monitor["name"])
		if gotName != "Empty Body" {
			t.Fatalf("expected unchanged name=%q, got %q", "Empty Body", gotName)
		}
	})

	t.Run("update_monitor_not_found", func(t *testing.T) {
		// Gate blocks with 403 since the consignee doesn't exist for this shop
		rec := h.doJSON(t, http.MethodPut, "/monitores/999999", &shopSession, map[string]any{
			"name": "Ghost",
		})
		assertStatus(t, rec, http.StatusForbidden)
	})

	t.Run("update_monitor_unauthorized", func(t *testing.T) {
		result := h.createMonitor(t, &shopSession, "Auth Test", fmt.Sprintf("api-upd-auth-%d@example.com", runID), "27998209673")

		// Second shop should not be able to edit first shop's monitor
		otherShopEmail := fmt.Sprintf("api-upd-other-shop-%d@example.com", runID)
		otherShopName := fmt.Sprintf("Outra Loja %d", runID)
		_ = h.createShop(t, &admin, otherShopName, otherShopEmail, "11988887777")
		otherSession := h.login(t, otherShopEmail, "11988887777")

		// Gate blocks with 403 since other shop doesn't own this monitor
		rec := h.doJSON(t, http.MethodPut, fmt.Sprintf("/monitores/%d", result.ID), &otherSession, map[string]any{
			"name": "Hacked",
		})
		assertStatus(t, rec, http.StatusForbidden)
	})
}

