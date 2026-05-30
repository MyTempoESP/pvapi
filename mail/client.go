package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	gomail "github.com/wneessen/go-mail"
)

// Mailer is the injectable interface for sending pickup confirmation emails.
type Mailer interface {
	SendPickupConfirmation(ctx context.Context, to string, data PickupEmailData) error
}

// Config holds SMTP server configuration.
type Config struct {
	Host       string
	Port       int
	User       string
	Pass       string
	From       string
	SkipVerify bool
}

// Client is the production Mailer backed by go-mail.
type Client struct {
	client *gomail.Client
	cfg    Config
}

// NewClient validates config and creates a go-mail client.
// Requires Host, User, Pass, and From.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("mail: SMTP host is required")
	}
	if cfg.User == "" || cfg.Pass == "" || cfg.From == "" {
		return nil, fmt.Errorf("mail: SMTP user, pass, and from are required")
	}

	opts := []gomail.Option{
		gomail.WithPort(cfg.Port),
		gomail.WithUsername(cfg.User),
		gomail.WithPassword(cfg.Pass),
		gomail.WithSMTPAuth(gomail.SMTPAuthLogin),
	}

	switch cfg.Port {
	case 465:
		opts = append(opts, gomail.WithSSL())
	case 587:
		opts = append(opts, gomail.WithTLSPolicy(gomail.TLSMandatory))
	default:
		opts = append(opts, gomail.WithTLSPolicy(gomail.NoTLS))
	}

	if cfg.SkipVerify {
		opts = append(opts, gomail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	}

	client, err := gomail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("mail: failed to create SMTP client: %w", err)
	}

	return &Client{client: client, cfg: cfg}, nil
}

// SendPickupConfirmation sends a pickup confirmation email.
// It renders the full template (order info + conditional login section).
func (c *Client) SendPickupConfirmation(
	ctx context.Context,
	to string,
	data PickupEmailData,
) error {
	html, err := renderPickupConfirmation(data)
	if err != nil {
		return fmt.Errorf("mail: failed to render pickup confirmation template: %w", err)
	}

	subject := fmt.Sprintf("Seu pedido #%s est\u00e1 pronto para retirada!", shortOrderID(data.OrderID))

	msg := gomail.NewMsg()
	if err := msg.From(c.cfg.From); err != nil {
		return fmt.Errorf("mail: invalid from address: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("mail: invalid to address: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextHTML, html)

	if data.QrCode != "" {
		qrPNG, err := decodeQrCodePNG(data.QrCode)
		if err != nil {
			return fmt.Errorf("mail: failed to decode qr code: %w", err)
		}
		if err := msg.EmbedReader("qrcode.png", bytes.NewReader(qrPNG)); err != nil {
			return fmt.Errorf("mail: failed to embed qr code: %w", err)
		}
	}

	slog.DebugContext(ctx, "sending pickup confirmation email",
		"to", to,
		"order_id", data.OrderID,
	)

	if err := c.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mail: failed to send pickup confirmation email to %s: %w", to, err)
	}

	slog.InfoContext(ctx, "pickup confirmation email sent",
		"to", to,
		"order_id", data.OrderID,
	)

	return nil
}

// decodeQrCodePNG extracts raw PNG bytes from a base64 data URI.
func decodeQrCodePNG(dataURI string) ([]byte, error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURI, prefix) {
		return nil, fmt.Errorf("mail: qr code data URI has unexpected prefix")
	}
	return base64.StdEncoding.DecodeString(dataURI[len(prefix):])
}

// NoopMailer is a no-op implementation of Mailer for tests.
type NoopMailer struct{}

func (NoopMailer) SendPickupConfirmation(_ context.Context, _ string, _ PickupEmailData) error {
	return nil
}

// SpyMailer records sent emails for test assertions.
type SpyMailer struct {
	SentConfirmation *PickupEmailData
}

func NewSpyMailer() *SpyMailer {
	return &SpyMailer{}
}

func (s *SpyMailer) SendPickupConfirmation(
	_ context.Context,
	_ string,
	data PickupEmailData,
) error {
	s.SentConfirmation = &data
	return nil
}
