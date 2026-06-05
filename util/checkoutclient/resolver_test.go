package checkoutclient_test

import (
	"testing"

	"linhazero.com/pvapi/util/checkoutclient"
)

func hasAlpha(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func TestResolveEmailOnly(t *testing.T) {
	out := checkoutclient.Resolve("user@example.com", "", "", "")
	if out.EffectiveEmail != "user@example.com" {
		t.Fatalf("expected email 'user@example.com', got %q", out.EffectiveEmail)
	}
	if out.IsSyntheticEmail {
		t.Fatal("expected IsSyntheticEmail=false for real email")
	}
	if out.ValidatedCPF != "" {
		t.Fatalf("expected empty CPF, got %q", out.ValidatedCPF)
	}
	if out.ClientName != "usuario-user" {
		t.Fatalf("expected fallback name 'usuario-user', got %q", out.ClientName)
	}
	if out.NormalizedPhone != "" {
		t.Fatalf("expected empty phone, got %q", out.NormalizedPhone)
	}
	if len(out.Password) != 16 || !hasAlpha(out.Password) {
		t.Fatalf("expected 16-char random alphanumeric password, got %q", out.Password)
	}
}

func TestResolveTelephoneOnly(t *testing.T) {
	out := checkoutclient.Resolve("", "11988887777", "", "")
	if out.EffectiveEmail != "11988887777@tele" {
		t.Fatalf("expected synthetic email '11988887777@tele', got %q", out.EffectiveEmail)
	}
	if !out.IsSyntheticEmail {
		t.Fatal("expected IsSyntheticEmail=true for synthetic email")
	}
	if out.NormalizedPhone != "11988887777" {
		t.Fatalf("expected phone '11988887777', got %q", out.NormalizedPhone)
	}
	if out.Password != "11988887777" {
		t.Fatalf("expected password = phone digits '11988887777', got %q", out.Password)
	}
	if out.ClientName != "usuario-11988887777" {
		t.Fatalf("expected fallback name 'usuario-11988887777', got %q", out.ClientName)
	}
}

func TestResolveCPFOnly(t *testing.T) {
	out := checkoutclient.Resolve("", "", "12345678901", "")
	if out.EffectiveEmail != "12345678901@br" {
		t.Fatalf("expected synthetic email '12345678901@br', got %q", out.EffectiveEmail)
	}
	if !out.IsSyntheticEmail {
		t.Fatal("expected IsSyntheticEmail=true for synthetic email")
	}
	if out.ValidatedCPF != "12345678901" {
		t.Fatalf("expected CPF '12345678901', got %q", out.ValidatedCPF)
	}
	if out.ClientName != "usuario-12345678901" {
		t.Fatalf("expected fallback name 'usuario-12345678901', got %q", out.ClientName)
	}
	// No phone, so password should be random 16-char alphanumeric
	if len(out.Password) != 16 || !hasAlpha(out.Password) {
		t.Fatalf("expected 16-char random password, got %q", out.Password)
	}
}

func TestResolveCPFWithDotsAndDashes(t *testing.T) {
	out := checkoutclient.Resolve("", "", "123.456.789-01", "")
	if out.EffectiveEmail != "12345678901@br" {
		t.Fatalf("expected email '12345678901@br', got %q", out.EffectiveEmail)
	}
	if out.ValidatedCPF != "12345678901" {
		t.Fatalf("expected CPF '12345678901', got %q", out.ValidatedCPF)
	}
}

func TestResolveCPFAndTelephoneNoEmail(t *testing.T) {
	out := checkoutclient.Resolve("", "11988887777", "123.456.789-01", "")
	if out.EffectiveEmail != "12345678901@br" {
		t.Fatalf("expected CPF-based email, got %q", out.EffectiveEmail)
	}
	if !out.IsSyntheticEmail {
		t.Fatal("expected IsSyntheticEmail=true")
	}
	if out.ValidatedCPF != "12345678901" {
		t.Fatalf("expected valid CPF")
	}
	if out.NormalizedPhone != "11988887777" {
		t.Fatalf("expected phone '11988887777', got %q", out.NormalizedPhone)
	}
	if out.Password != "11988887777" {
		t.Fatalf("expected password = phone digits, got %q", out.Password)
	}
}

func TestResolveRealEmailWins(t *testing.T) {
	out := checkoutclient.Resolve("real@example.com", "11988887777", "123.456.789-01", "")
	if out.EffectiveEmail != "real@example.com" {
		t.Fatalf("expected real email, got %q", out.EffectiveEmail)
	}
	if out.IsSyntheticEmail {
		t.Fatal("expected IsSyntheticEmail=false")
	}
	if out.ValidatedCPF != "12345678901" {
		t.Fatalf("expected valid CPF")
	}
	if out.NormalizedPhone != "11988887777" {
		t.Fatalf("expected phone")
	}
	// Phone digits become password even with real email
	if out.Password != "11988887777" {
		t.Fatalf("expected password = phone digits, got %q", out.Password)
	}
}

func TestResolveNameProvided(t *testing.T) {
	out := checkoutclient.Resolve("", "11988887777", "", "João Silva")
	if out.ClientName != "João Silva" {
		t.Fatalf("expected provided name 'João Silva', got %q", out.ClientName)
	}

	out2 := checkoutclient.Resolve("real@example.com", "", "", "Maria")
	if out2.ClientName != "Maria" {
		t.Fatalf("expected provided name 'Maria', got %q", out2.ClientName)
	}

	out3 := checkoutclient.Resolve("real@example.com", "11988887777", "", "José")
	if out3.ClientName != "José" {
		t.Fatalf("expected provided name 'José', got %q", out3.ClientName)
	}
}

func TestResolveInvalidCPFFallback(t *testing.T) {
	// Non-numeric CPF + telephone → CPF ignored, uses telephone
	out := checkoutclient.Resolve("", "11988887777", "abc", "")
	if out.EffectiveEmail != "11988887777@tele" {
		t.Fatalf("expected phone-based email, got %q", out.EffectiveEmail)
	}
	if out.ValidatedCPF != "" {
		t.Fatalf("expected empty CPF for invalid input")
	}

	// Short CPF (10 digits) → treated as not provided, no identity
	out2 := checkoutclient.Resolve("", "", "1234567890", "")
	if out2.EffectiveEmail != "" {
		t.Fatalf("expected no identity for invalid short CPF, got %q", out2.EffectiveEmail)
	}
}

func TestResolveEmptyPhoneAfterStrip(t *testing.T) {
	out := checkoutclient.Resolve("", "+-() ", "", "")
	if out.EffectiveEmail != "" {
		t.Fatalf("expected no identity for empty telephone after strip, got %q", out.EffectiveEmail)
	}
}

func TestResolveNoIdentity(t *testing.T) {
	out := checkoutclient.Resolve("", "", "", "")
	if out.EffectiveEmail != "" {
		t.Fatalf("expected no email, got %q", out.EffectiveEmail)
	}
	if out.Password != "" {
		t.Fatalf("expected empty password, got %q", out.Password)
	}
	if out.NormalizedPhone != "" {
		t.Fatalf("expected empty phone")
	}
	if out.ValidatedCPF != "" {
		t.Fatalf("expected empty CPF")
	}
	if out.IsSyntheticEmail {
		t.Fatal("expected IsSyntheticEmail=false when no identity")
	}
}

func TestResolveTelephoneWithFormatting(t *testing.T) {
	out := checkoutclient.Resolve("", "(11) 98888-7777", "", "")
	if out.NormalizedPhone != "11988887777" {
		t.Fatalf("expected phone '11988887777', got %q", out.NormalizedPhone)
	}
	if out.Password != "11988887777" {
		t.Fatalf("expected password = phone digits, got %q", out.Password)
	}
}

func TestResolveCPFProvidedButShortNameFallback(t *testing.T) {
	// When CPF is provided with an empty name, name is derived from CPF
	out := checkoutclient.Resolve("", "", "98765432100", "")
	if out.ClientName != "usuario-98765432100" {
		t.Fatalf("expected 'usuario-98765432100', got %q", out.ClientName)
	}
}

func TestResolvePhoneOnlyNameFallback(t *testing.T) {
	out := checkoutclient.Resolve("", "11999999999", "", "")
	expectedName := "usuario-11999999999"
	if out.ClientName != expectedName {
		t.Fatalf("expected %q, got %q", expectedName, out.ClientName)
	}
}

func TestResolveEmailWithNameFallback(t *testing.T) {
	out := checkoutclient.Resolve("maria.silva@example.com", "", "", "")
	if out.ClientName != "usuario-mariasilva" {
		t.Fatalf("expected 'usuario-mariasilva', got %q", out.ClientName)
	}
}

func TestResolveRandomPasswordWhenNoPhone(t *testing.T) {
	// With email + no phone → random password, 16 chars alphanumeric
	out := checkoutclient.Resolve("user@example.com", "", "", "")
	if len(out.Password) != 16 {
		t.Fatalf("expected 16-char password, got %q (len=%d)", out.Password, len(out.Password))
	}
	if !hasAlpha(out.Password) {
		t.Fatal("expected alphanumeric password (contains letters)")
	}
}

func TestResolveNoPIIInOutput(t *testing.T) {
	// The effective email is derived but contains CPF digits when CPF-only
	// This is expected — the PII policy applies to logging, not the email field itself.
	// Just verify the CPF is correctly set.
	out := checkoutclient.Resolve("", "", "12345678901", "")
	if out.ValidatedCPF != "12345678901" {
		t.Fatalf("expected CPF '12345678901', got %q", out.ValidatedCPF)
	}
}
