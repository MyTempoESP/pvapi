---
# pvapi-63xz
title: Welcome email via go-mail on client account creation
status: completed
type: feature
priority: normal
created_at: 2026-05-29T23:12:39Z
updated_at: 2026-05-29T23:33:10Z
blocked_by:
    - pvapi-7vz8
---

## Problem Statement

When a client account is auto-created during checkout, the buyer receives no notification — they don't know their account exists, what their temporary password is, or how to log in. The back-end has no email infrastructure at all. The existing `docs/email.html` is a PHP template for password reset, not a welcome email, and does not work with Go.

## Solution

Integrate `github.com/wneessen/go-mail` to send a welcome email to the buyer when their client account is created during checkout. A new `mail/` package encapsulates SMTP configuration, template rendering, and sending. The `docs/email.html` template is rewritten as a Go `html/template` containing the buyer's login email, temporary password, login link, and password-reset link. Email sending is best-effort — it runs after checkout completes, the response is already sent, and failures are logged without affecting the purchase.

SMTP configuration comes from environment variables. If not configured, email sending is silently skipped. Partial configuration is treated as an error.

## User Stories

1. As a buyer whose account was auto-created, I want to receive a welcome email with my login credentials, so I know my account exists and how to access it.
2. As a buyer, I want the email to include a login link, so I can reach the storefront without typing the URL.
3. As a buyer, I want the email to include a password-reset link, so I can change my temporary password securely.
4. As a storefront operator, I want the PVAuto brand identity reflected in the email template, so the email looks professional and trustworthy.
5. As a developer, I want SMTP configuration isolated in environment variables, so I can change providers without touching code.
6. As a developer, I want the mailer to be injectable, so I can test the checkout controller without a real SMTP server.
7. As a developer, I want the email template embedded in the binary, so there is no runtime filesystem dependency.
8. As an operator, I want the server to boot cleanly when SMTP is not configured, so email can be deferred to a later deployment.
9. As an operator, I want TLS verification skippable for local/dev SMTP servers, so I can test with Mailpit/MailHog.
10. As an operator, I want clearly distinguishable log lines for email successes and failures, so I can monitor delivery in production.

## Implementation Decisions

**New `mail/` package**

A top-level `mail/` package with two files:
- One file for SMTP client setup and sending. Exposes `Config` (host, port, user, pass, from, skipVerify), `NewClient(cfg Config) (*Client, error)`, and `SendWelcome(ctx, email, name, password, loginURL, resetURL)`.
- One file for the email template, using `//go:embed docs/email.html` to ship the template inside the binary.

**Dependency: `github.com/wneessen/go-mail`**

Used for SMTP dialing, message construction, and sending. TLS policy is derived from port: 465 → implicit TLS, 587 → STARTTLS, other → no TLS. `PV_API_SMTP_SKIP_VERIFY` (boolean env) skips certificate verification.

**Environment variables (all in `env/env.go`)**

| Variable | Required | Purpose |
|---|---|---|
| `PV_API_SMTP_HOST` | Yes (if sending) | SMTP server hostname |
| `PV_API_SMTP_PORT` | Yes (if sending) | SMTP port |
| `PV_API_SMTP_USER` | Yes (if sending) | SMTP authentication username |
| `PV_API_SMTP_PASS` | Yes (if sending) | SMTP authentication password |
| `PV_API_SMTP_FROM` | Yes (if sending) | From: address for outgoing mail |
| `PV_API_SMTP_SKIP_VERIFY` | No | Skip TLS certificate verification |
| `PV_API_LOGIN_LINK` | Yes (if sending) | URL for the login button in emails |
| `PV_API_PASSWORD_RESET_LINK` | Yes (if sending) | URL for the password-reset link in emails |

Configuration validation: if `PV_API_SMTP_HOST` is empty, mail is silently disabled (no error). If `PV_API_SMTP_HOST` is set but any required sibling is missing, `NewClient` returns an error logged at startup.

**Email template**

`docs/email.html` rewritten as a `html/template`-compatible Go template. Template data:

```
Name      string
Email     string
Password  string
LoginURL  string
ResetURL  string
```

The email greets the buyer by name, displays their login email and temporary password inline, and provides two action links: login and password reset. Branding matches the existing PVAuto identity (colors, logo space, font stack).

**Injectability**

The checkout controller accepts a mailer interface. In production, a real `mail.Client` is wired in. In tests, a no-op implementation is used. The mailer is called after the HTTP response is sent, best-effort only — failures are logged at `slog.Error` level.

## Testing Decisions

Two testing layers:

1. **Template unit test**: Parse the embedded template with sample data, render to string, assert key strings appear (login:, the email value, the password value, the login link, the reset link). No network required.
2. **Controller integration test**: Pass a spy mailer to the checkout controller. Assert that when `client_email` is present and checkout succeeds, the mailer's `SendWelcome` is called exactly once with the correct recipient. No real SMTP required.

Good templates tests are pure — they assert the rendered output contains what matters, not the exact HTML structure.

## Out of Scope

- The client account auto-creation flow itself (covered by PRD "Client account auto-creation during checkout"). This PRD depends on that one.
- Password reset email — only the welcome email for newly created accounts.
- Email queueing, retry, or async delivery infrastructure.
- HTML email design — the template uses the existing PVAuto design language.
- Multi-language emails.
- Email open/click tracking.
- Rate limiting on email sends.

## Further Notes

This PRD is blocked by "Client account auto-creation during checkout". The mail package is designed to be called from the checkout controller at the point where `CreateAndAssociate` succeeds. The package is self-contained and could be reused later for password-reset emails or other transactional notifications.



Triage: ready-for-agent



## Summary of Changes

### New files
- **mail/client.go**: Mailer interface, Config struct, Client (production implementation backed by go-mail), NoopMailer, SpyMailer, NewClient.
- **mail/template.go**: Go html/template for the welcome email, embedded via //go:embed welcome.html.
- **mail/welcome.html**: HTML email template for welcome emails (PVAuto-branded, Portuguese).
- **mail/mail_test.go**: Unit tests for template rendering (name, email, password, URLs, XSS escaping), NoopMailer and SpyMailer.

### Modified files

- **env/env.go**: Added SMTP environment variables: HOST, PORT, USER, PASS, FROM, LOGIN_LINK, PASSWORD_RESET_LINK, SKIP_VERIFY (boolean).

- **docs/email.html**: Rewritten as a Go html/template with welcome-email content (greeting, login credentials, login/reset action links). Removed PHP-specific syntax.

- **controllers/checkout/checkout.go**: 
  - ConfirmHandler now accepts a mail.Mailer parameter.
  - createClientAfterCheckout accepts and calls the mailer after client creation succeeds (best-effort, logs errors).
  - Reads PV_API_LOGIN_LINK and PV_API_PASSWORD_RESET_LINK from env. Skips sending if either is empty.

- **routes.go**: setupRoutes accepts a mail.Mailer parameter, passes it to the checkout handler.

- **main.go**: 
  - Added createMailer function that reads SMTP env vars, validates config, and creates a mail.Client (returns nil if SMTP_HOST unset — mail silently disabled).
  - Passes mailer to setupRoutes.

- **api_harness_test.go**: New apiHarness.Mailer field (SpyMailer), passed to setupRoutes in newAPIHarness.

- **api_checkout_client_test.go**: Added two test cases:
  - welcome_email_sent_on_client_creation — verifies SpyMailer records exactly one email with correct recipient, name, login URL, and reset URL.
  - no_welcome_email_without_client_email — verifies no email sent when client_email is absent.

### go.mod
- Added dependency: github.com/wneessen/go-mail v0.6.0

### TLS policy
- Port 465 → implicit SSL/TLS (WithSSL)
- Port 587 → STARTTLS (WithTLSPolicy TLSMandatory)
- Other ports → no TLS
- PV_API_SMTP_SKIP_VERIFY → TLS config with InsecureSkipVerify
