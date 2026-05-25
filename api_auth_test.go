package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAuth(t *testing.T) {
	h := newAPIHarness(t)

	runID := time.Now().UnixNano()
	adminEmail := "sistemas@lzero.org"
	registerEmail := fmt.Sprintf("api-auth-%d@example.com", runID)
	registerPassword := "abc123"

	t.Run("unauthenticated_user", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet, "/user", nil, "", nil)
		assertStatus(t, rec, http.StatusUnauthorized)
	})

	t.Run("register", func(t *testing.T) {
		rec := h.doJSON(t, http.MethodPost, "/register", nil, map[string]any{
			"name":     "api user",
			"email":    registerEmail,
			"password": registerPassword,
		})
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("login", func(t *testing.T) {
		session := h.login(t, registerEmail, registerPassword)
		if len(session.cookies) == 0 {
			t.Fatal("expected auth cookies after login")
		}
	})

	t.Run("logout", func(t *testing.T) {
		session := h.login(t, registerEmail, registerPassword)
		rec := h.doRequest(t, http.MethodGet, "/logout", &session, "", nil)
		assertStatus(t, rec, http.StatusOK)
	})

	t.Run("public_permissions", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/permissions?email=%s", registerEmail),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["data"])
		if report["highest"] != "user" {
			t.Fatalf("expected highest 'user', got %v", report["highest"])
		}
	})

	t.Run("admin_permissions", func(t *testing.T) {
		rec := h.doRequest(t, http.MethodGet,
			fmt.Sprintf("/permissions?email=%s", adminEmail),
			nil, "", nil)
		assertStatus(t, rec, http.StatusOK)

		body := decodeJSON(t, rec)
		report := asMap(t, body["data"])
		if report["highest"] != "admin" {
			t.Fatalf("expected highest 'admin', got %v", report["highest"])
		}
		if _, exists := report["admin"]; !exists {
			t.Fatal("expected 'admin' role to exist")
		}
	})
}

func TestAuthAdminLogin(t *testing.T) {
	h := newAPIHarness(t)
	_ = h.login(t, "sistemas@lzero.org", "TestingPassword")
}
