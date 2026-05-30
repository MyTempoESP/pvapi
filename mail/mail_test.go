package mail

import (
	"context"
	"strings"
	"testing"
)

func TestRenderPickupConfirmation(t *testing.T) {
	data := PickupEmailData{
		OrderID:   "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		QrCode:    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		IsNewUser: true,
		Name:      "João Silva",
		Email:     "joao@example.com",
		Password:  "A3bC9xYz",
		LoginURL:  "https://pvaauto.example.com/login",
		ResetURL:  "https://pvaauto.example.com/reset",
		Groups: []PickupGroup{
			{
				PickupName: "Loja Centro",
				Items: []PickupItem{
					{ProductName: "Produto A", Quantity: 2},
				},
			},
		},
	}

	html, err := renderPickupConfirmation(data)
	if err != nil {
		t.Fatalf("renderPickupConfirmation failed: %v", err)
	}

	cases := []struct {
		name string
		want string
	}{
		{"includes order ID (short)", "a1b2c3d4"},
		{"includes order ID (full)", "a1b2c3d4-e5f6-7890-abcd-ef1234567890"},
		{"includes QR code as CID image", "cid:qrcode.png"},
		{"includes pickup location name", "Loja Centro"},
		{"includes product name", "Produto A"},
		{"includes product quantity", "x2"},
		{"greets by name", "João Silva"},
		{"includes email", "joao@example.com"},
		{"includes password (new user)", "A3bC9xYz"},
		{"includes login link", "https://pvaauto.example.com/login"},
		{"includes reset link", "https://pvaauto.example.com/reset"},
		{"has Portuguese subject heading", "Pedido"},
		{"has confirmation text", "retirada"},
		{"has configure account button", "Configurar minha conta"},
		{"has reset link text", "Redefinir senha"},
		{"has subject line reference with order short", "#a1b2c3d4"},
		{"has QR instructional text", "Apresente este QR code no balcão"},
		{"has logo URL", "https://pvauto.online/logo.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(html, tc.want) {
				t.Errorf("expected rendered template to contain %q", tc.want)
			}
		})
	}
}

func TestRenderPickupConfirmationExistingUser(t *testing.T) {
	data := PickupEmailData{
		OrderID:   "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		QrCode:    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		IsNewUser: false,
		Name:      "Maria Oliveira",
		Email:     "maria@example.com",
		Groups: []PickupGroup{
			{
				PickupName: "Loja Norte",
				Items: []PickupItem{
					{ProductName: "Produto B", Quantity: 1},
				},
			},
		},
	}

	html, err := renderPickupConfirmation(data)
	if err != nil {
		t.Fatalf("renderPickupConfirmation failed: %v", err)
	}

	// Existing user should see order info but no login credentials
	loginSections := []string{
		"Senha temporária",
		"Configurar minha conta",
		"Redefinir senha",
	}
	for _, section := range loginSections {
		if strings.Contains(html, section) {
			t.Errorf("existing user template should not contain %q", section)
		}
	}

	// Should still show order info
	if !strings.Contains(html, "b2c3d4e5") {
		t.Errorf("expected order ID short in template")
	}
	if !strings.Contains(html, "Loja Norte") {
		t.Errorf("expected pickup location in template")
	}
	if !strings.Contains(html, "Produto B") {
		t.Errorf("expected product name in template")
	}
}

func TestRenderPickupConfirmationMultipleGroups(t *testing.T) {
	data := PickupEmailData{
		OrderID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
		IsNewUser: false,
		Groups: []PickupGroup{
			{
				PickupName: "Loja Centro",
				Items: []PickupItem{
					{ProductName: "Produto A", Quantity: 1},
				},
			},
			{
				PickupName: "Loja Norte",
				Items: []PickupItem{
					{ProductName: "Produto B", Quantity: 3},
					{ProductName: "Produto C", Quantity: 2},
				},
			},
		},
	}

	html, err := renderPickupConfirmation(data)
	if err != nil {
		t.Fatalf("renderPickupConfirmation failed: %v", err)
	}

	if !strings.Contains(html, "Loja Centro") {
		t.Errorf("expected first pickup location")
	}
	if !strings.Contains(html, "Loja Norte") {
		t.Errorf("expected second pickup location")
	}
	if !strings.Contains(html, "Produto A") {
		t.Errorf("expected first product")
	}
	if !strings.Contains(html, "Produto B") {
		t.Errorf("expected second product")
	}
	if !strings.Contains(html, "Produto C") {
		t.Errorf("expected third product")
	}
}

func TestRenderPickupConfirmationHTMLEscaped(t *testing.T) {
	data := PickupEmailData{
		OrderID:   "d4e5f6a7-b8c9-0123-defa-234567890123",
		QrCode:    "data:image/png;base64,iVBOR",
		IsNewUser: true,
		Name:      "João <script>alert('xss')</script> Silva",
		Email:     "joao@example.com",
		Password:  "pass123",
		LoginURL:  "https://example.com/login",
		ResetURL:  "https://example.com/reset",
		Groups: []PickupGroup{
			{
				PickupName: "Loja Teste",
				Items: []PickupItem{
					{ProductName: "Produto <b>Teste</b>", Quantity: 1},
				},
			},
		},
	}

	html, err := renderPickupConfirmation(data)
	if err != nil {
		t.Fatalf("renderPickupConfirmation failed: %v", err)
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
	err := m.SendPickupConfirmation(context.Background(), "a@b.com", PickupEmailData{OrderID: "abc"})
	if err != nil {
		t.Errorf("NoopMailer should not return an error, got: %v", err)
	}
}

func TestSpyMailer(t *testing.T) {
	spy := NewSpyMailer()

	if spy.SentConfirmation != nil {
		t.Fatalf("expected no sent email initially, got %v", spy.SentConfirmation)
	}

	data := PickupEmailData{
		OrderID:   "e5f6a7b8-c9d0-1234-efab-345678901234",
		QrCode:    "data:image/png;base64,abc123",
		IsNewUser: true,
		Name:      "User Test",
		Email:     "user@test.com",
		Password:  "pass123",
		LoginURL:  "https://login.url",
		ResetURL:  "https://reset.url",
		Groups: []PickupGroup{
			{
				PickupName: "Pickup A",
				Items: []PickupItem{
					{ProductName: "Product X", Quantity: 2},
				},
			},
		},
	}

	err := spy.SendPickupConfirmation(context.Background(), "user@test.com", data)
	if err != nil {
		t.Fatalf("SpyMailer.SendPickupConfirmation failed: %v", err)
	}

	if spy.SentConfirmation == nil {
		t.Fatal("expected SentConfirmation to be set")
	}

	if spy.SentConfirmation.OrderID != "e5f6a7b8-c9d0-1234-efab-345678901234" {
		t.Errorf("expected OrderID, got %q", spy.SentConfirmation.OrderID)
	}
	if spy.SentConfirmation.IsNewUser != true {
		t.Errorf("expected IsNewUser=true, got %v", spy.SentConfirmation.IsNewUser)
	}
	if spy.SentConfirmation.Email != "user@test.com" {
		t.Errorf("expected Email=user@test.com, got %q", spy.SentConfirmation.Email)
	}
	if spy.SentConfirmation.Password != "pass123" {
		t.Errorf("expected Password=pass123, got %q", spy.SentConfirmation.Password)
	}
	if spy.SentConfirmation.LoginURL != "https://login.url" {
		t.Errorf("expected LoginURL, got %q", spy.SentConfirmation.LoginURL)
	}
	if spy.SentConfirmation.ResetURL != "https://reset.url" {
		t.Errorf("expected ResetURL, got %q", spy.SentConfirmation.ResetURL)
	}
	if len(spy.SentConfirmation.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(spy.SentConfirmation.Groups))
	}
	if spy.SentConfirmation.Groups[0].PickupName != "Pickup A" {
		t.Errorf("expected PickupName=Pickup A, got %q", spy.SentConfirmation.Groups[0].PickupName)
	}
	if len(spy.SentConfirmation.Groups[0].Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(spy.SentConfirmation.Groups[0].Items))
	}
	if spy.SentConfirmation.Groups[0].Items[0].ProductName != "Product X" {
		t.Errorf("expected ProductName=Product X, got %q", spy.SentConfirmation.Groups[0].Items[0].ProductName)
	}
}
