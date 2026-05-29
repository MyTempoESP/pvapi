package mail

import (
	"context"
	"strings"
	"testing"
)

func TestRenderWelcome(t *testing.T) {
	loginURL := "https://pvaauto.example.com/login"
	resetURL := "https://pvaauto.example.com/reset"

	html, err := renderWelcome(
		"João Silva",
		"joao@example.com",
		"A3bC9xYz",
		loginURL,
		resetURL,
	)
	if err != nil {
		t.Fatalf("renderWelcome failed: %v", err)
	}

	cases := []struct {
		name string
		want string
	}{
		{"greets by name", "João Silva"},
		{"includes email", "joao@example.com"},
		{"includes password", "A3bC9xYz"},
		{"includes login link", "https://pvaauto.example.com/login"},
		{"includes reset link", "https://pvaauto.example.com/reset"},
		{"has Portuguese heading", "Bem-vindo"},
		{"has login button text", "Acessar minha conta"},
		{"has reset link text", "Redefinir senha"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(html, tc.want) {
				t.Errorf("expected rendered template to contain %q", tc.want)
			}
		})
	}
}

func TestRenderWelcomeHTMLEscaped(t *testing.T) {
	// html/template auto-escapes — name with special chars should be escaped
	html, err := renderWelcome(
		"João <script>alert('xss')</script> Silva",
		"joao@example.com",
		"pass123",
		"https://example.com/login",
		"https://example.com/reset",
	)
	if err != nil {
		t.Fatalf("renderWelcome failed: %v", err)
	}

	if strings.Contains(html, "<script>") {
		t.Errorf("expected HTML-escaped output, but raw <script> tag found")
	}

	if !strings.Contains(html, "João") {
		t.Errorf("expected name to appear (unescaped unicode)")
	}
}

func TestNoopMailer(t *testing.T) {
	m := NoopMailer{}
	err := m.SendWelcome(context.Background(), "a@b.com", "n", "p", "l", "r")
	if err != nil {
		t.Errorf("NoopMailer should not return an error, got: %v", err)
	}
}

func TestSpyMailer(t *testing.T) {
	spy := NewSpyMailer()

	if len(spy.SentEmails) != 0 {
		t.Fatalf("expected no sent emails initially, got %d", len(spy.SentEmails))
	}

	err := spy.SendWelcome(context.Background(), "user@test.com", "User", "pass123", "https://login.url", "https://reset.url")
	if err != nil {
		t.Fatalf("SpyMailer.SendWelcome failed: %v", err)
	}

	if len(spy.SentEmails) != 1 {
		t.Fatalf("expected 1 sent email, got %d", len(spy.SentEmails))
	}

	email := spy.SentEmails[0]
	if email.To != "user@test.com" {
		t.Errorf("expected To=user@test.com, got %q", email.To)
	}
	if email.Name != "User" {
		t.Errorf("expected Name=User, got %q", email.Name)
	}
	if email.Password != "pass123" {
		t.Errorf("expected Password=pass123, got %q", email.Password)
	}
	if email.LoginURL != "https://login.url" {
		t.Errorf("expected LoginURL=https://login.url, got %q", email.LoginURL)
	}
	if email.ResetURL != "https://reset.url" {
		t.Errorf("expected ResetURL=https://reset.url, got %q", email.ResetURL)
	}
}
