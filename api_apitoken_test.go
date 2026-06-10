package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"linhazero.com/pvapi/models/apitokenmodel"
)

func TestAPITokenAuth(t *testing.T) {
	h := newAPIHarness(t)

	t.Run("missing_header_returns_401", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", nil, map[string]any{
			"value_paid": "3.00",
			"items": []map[string]any{},
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("valid_token_succeeds", func(t *testing.T) {
		// The harness already created an API client
		// Expect 400 (validation error for empty items), not 401 — auth passed
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", h.defaultAPISession, map[string]any{
			"value_paid": "3.00",
			"items": []map[string]any{},
		})
		assertStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("garbage_token_returns_401", func(t *testing.T) {
		session := &apiSession{bearerToken: "pv_" + strings.Repeat("a", 64)}
		rec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", session, map[string]any{
			"value_paid": "3.00",
			"items": []map[string]any{},
		})
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("admin_create_api_client", func(t *testing.T) {
		admin := h.login(t, "sistemas@lzero.org", "TestingPassword")

		rec := h.doJSON(t, http.MethodPost, "/admin/api-clients", &admin, map[string]any{
			"name":  "Test Client via Admin",
			"email": fmt.Sprintf("admin-created-%d@test.com", time.Now().UnixNano()),
		})
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		if asString(t, body["message"]) != "API client created" {
			t.Fatalf("expected 'API client created', got: %v", body["message"])
		}

		// token should be returned
		token := asString(t, body["token"])
		if !strings.HasPrefix(token, "pv_") {
			t.Fatalf("expected token starting with pv_, got: %s", token)
		}
	})

	t.Run("admin_list_api_clients", func(t *testing.T) {
		admin := h.login(t, "sistemas@lzero.org", "TestingPassword")

		rec := h.doJSON(t, http.MethodGet, "/admin/api-clients", &admin, nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("admin_revoke_api_client", func(t *testing.T) {
		admin := h.login(t, "sistemas@lzero.org", "TestingPassword")

		// Create a new client first
		createRec := h.doJSON(t, http.MethodPost, "/admin/api-clients", &admin, map[string]any{
			"name":  "Revokable Client",
			"email": fmt.Sprintf("revokable-%d@test.com", time.Now().UnixNano()),
		})
		assertStatus(t, createRec, http.StatusOK)
		createBody := decodeJSON(t, createRec)
		clientID := asInt(t, asMap(t, createBody["api_client"])["id"])

		// Revoke it
		revokeRec := h.doJSON(t, http.MethodDelete, fmt.Sprintf("/admin/api-clients/%d", clientID), &admin, nil)
		assertStatus(t, revokeRec, http.StatusOK)

		// The token should no longer work
		token := asString(t, createBody["token"])
		revokedSession := &apiSession{bearerToken: token}
		chkRec := h.doJSON(t, http.MethodPost, "/checkouts/confirm", revokedSession, map[string]any{
			"value_paid": "3.00",
			"items": []map[string]any{},
		})
		assertStatus(t, chkRec, http.StatusUnauthorized)
		body := decodeJSON(t, chkRec)
		if asString(t, body["message"]) != "token revoked" {
			t.Fatalf("expected 'token revoked', got: %v", body["message"])
		}
	})
}

func TestAPITokenModel(t *testing.T) {
	h := newAPIHarness(t)

	t.Run("token_format", func(t *testing.T) {
		token, err := apitokenmodel.GenerateToken()
		if err != nil {
			t.Fatalf("GenerateToken: %v", err)
		}
		if !strings.HasPrefix(token, "pv_") {
			t.Fatalf("expected pv_ prefix, got: %s", token)
		}
		if len(token) != 3+64 { // pv_ + 64 hex chars
			t.Fatalf("expected len %d, got %d: %s", 3+64, len(token), token)
		}
	})

	t.Run("hash_deterministic", func(t *testing.T) {
		token := "pv_abc123"
		h1 := apitokenmodel.HashToken(token)
		h2 := apitokenmodel.HashToken(token)
		if h1 != h2 {
			t.Fatalf("hash should be deterministic: %s != %s", h1, h2)
		}
	})

	t.Run("round_trip", func(t *testing.T) {
		id, token, err := apitokenmodel.Create(context.Background(), h.pool, "Test", "test@test.com")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if id == 0 {
			t.Fatal("expected non-zero id")
		}

		result, err := apitokenmodel.ValidateToken(context.Background(), h.pool, token)
		if err != nil {
			t.Fatalf("ValidateToken: %v", err)
		}
		if result.ID != id {
			t.Fatalf("expected id %d, got %d", id, result.ID)
		}
	})

	t.Run("revoked_token_rejected", func(t *testing.T) {
		id, token, err := apitokenmodel.Create(context.Background(), h.pool, "Revoke Test", "revoke@test.com")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		if err := apitokenmodel.Revoke(context.Background(), h.pool, id); err != nil {
			t.Fatalf("Revoke: %v", err)
		}

		_, err = apitokenmodel.ValidateToken(context.Background(), h.pool, token)
		if err != apitokenmodel.ErrTokenRevoked {
			t.Fatalf("expected ErrTokenRevoked, got: %v", err)
		}
	})

	t.Run("garbage_token_not_found", func(t *testing.T) {
		_, err := apitokenmodel.ValidateToken(context.Background(), h.pool, "pv_nonexistent")
		if err != apitokenmodel.ErrTokenNotFound {
			t.Fatalf("expected ErrTokenNotFound, got: %v", err)
		}
	})
}

