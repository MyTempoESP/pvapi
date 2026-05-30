package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	gomail "github.com/wneessen/go-mail"
)

// Mailer is the injectable interface for sending welcome emails.
type Mailer interface {
	SendWelcome(ctx context.Context, to, name, password, loginURL, resetURL string) error
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

// SendWelcome sends a welcome email with login credentials.
func (c *Client) SendWelcome(
	ctx context.Context,
	to, name, password, loginURL, resetURL string,
) error {
	html, err := renderWelcome(name, to, password, loginURL, resetURL)
	if err != nil {
		return fmt.Errorf("mail: failed to render welcome template: %w", err)
	}

	msg := gomail.NewMsg()
	if err := msg.From(c.cfg.From); err != nil {
		return fmt.Errorf("mail: invalid from address: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("mail: invalid to address: %w", err)
	}
	msg.Subject("Bem-vindo à PVAuto!")
	msg.SetBodyString(gomail.TypeTextHTML, html)

	slog.DebugContext(ctx, "sending welcome email",
		"to", to,
	)

	if err := c.client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("mail: failed to send welcome email to %s: %w", to, err)
	}

	slog.InfoContext(ctx, "welcome email sent",
		"to", to,
	)

	return nil
}

// NoopMailer is a no-op implementation of Mailer for tests.
type NoopMailer struct{}

func (NoopMailer) SendWelcome(_ context.Context, _, _, _, _, _ string) error {
	return nil
}

// SpyMailer records sent emails for test assertions.
type SpyMailer struct {
	SentEmails []WelcomeEmail
}

type WelcomeEmail struct {
	To       string
	Name     string
	Password string
	LoginURL string
	ResetURL string
}

func NewSpyMailer() *SpyMailer {
	return &SpyMailer{SentEmails: make([]WelcomeEmail, 0)}
}

func (s *SpyMailer) SendWelcome(
	_ context.Context,
	to, name, password, loginURL, resetURL string,
) error {
	s.SentEmails = append(s.SentEmails, WelcomeEmail{
		To:       to,
		Name:     name,
		Password: password,
		LoginURL: loginURL,
		ResetURL: resetURL,
	})
	return nil
}
