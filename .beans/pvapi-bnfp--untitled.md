---
# pvapi-bnfp
title: 'Refactor: pickup confirmation email with conditional login details'
status: done
type: task
priority: normal
created_at: 2026-05-30T21:41:06Z
updated_at: 2026-05-30T21:49:02Z
---

## Problem Statement

After checkout, customers receive a welcome email containing only login credentials (email + temporary password). They receive no order confirmation, no pickup instructions, and no QR code. The email serves a single purpose — account creation — while ignoring the primary user need: knowing their order details and how to collect it.

Existing customers who check out again receive the same welcome email with fresh credentials they don't need, since their account already exists.

## Solution

Replace the welcome email with a pickup confirmation email that always includes order information (order ID, items, pickup locations, QR code) and only conditionally includes login details when the checkout creates a new user account.

The email template has a fixed order-info section and an optional login section gated by a boolean flag. The mail package exposes a single public method (`SendPickupConfirmation`) that internally dispatches to the appropriate internal renderer based on whether the user is new.

## User Stories

1. As a customer who just checked out, I want to receive an email with my order number, items, and pickup location, so that I know exactly where and when to collect my products.

2. As a customer who just checked out, I want a scannable QR code in my confirmation email, so that I can present it at the warehouse counter for quick retrieval.

3. As a first-time customer, I want to receive login credentials alongside my order confirmation, so that I can access my account dashboard without a separate onboarding flow.

4. As a returning customer, I want to receive only order information (no redundant credentials), so that my inbox stays clean and I don't receive passwords for an account I already have.

5. As a customer purchasing items from multiple pickup locations, I want my confirmation email to group items by location, so that I know which products to collect at each place.

6. As a customer viewing the email on mobile, I want the truncated order ID in the subject line, so that I can quickly identify the right email in my inbox.

7. As a customer, I want the order ID in the email subject to be visually styled as a code reference, so that it stands out from the surrounding text.

8. As an operator reviewing email templates, I want the template to use HTML conditional blocks rather than separate template files, so that the order-info layout is maintained in one place.

9. As a developer testing the email flow, I want the spy mailer to record the full email data struct, so that I can assert on both order info and login details in integration tests.

10. As a returning customer with the same email across multiple checkouts, I want the system to detect my existing account and skip credential generation, so that my account password is never accidentally overwritten.

## Implementation Decisions

### Mail interface rename and signature

The `Mailer` interface method `SendWelcome` is replaced with `SendPickupConfirmation(ctx, to string, data PickupEmailData) error`. All implementations (`Client`, `NoopMailer`, `SpyMailer`) adopt the new method and drop the old one.

### Single template with conditional login block

The HTML template (`pickup_confirmation.html`) uses Go template conditionals (`{{if .IsNewUser}}`) to optionally render the login section. The order-info section is always rendered. This avoids duplicating the order layout across two templates.

### Orchestrator pattern in the mail package

`SendPickupConfirmation` is the single public entry point. It inspects `data.IsNewUser` and delegates to one of two internal render+send methods: `sendPickupOnly` (order info only) or `sendPickupWithLogin` (order info + login details). The caller never chooses between them.

### PickupEmailData shape (from prototype)

```go
type PickupItem struct {
    ProductName string
    Quantity    int
}

type PickupGroup struct {
    PickupName string
    Items      []PickupItem
}

type PickupEmailData struct {
    OrderID   string
    QrCode    string
    Groups    []PickupGroup
    IsNewUser bool
    Name      string
    Email     string
    Password  string
    LoginURL  string
    ResetURL  string
}
```

Items are grouped hierarchically by pickup location. The template iterates groups → items, rendering section headers per location.

### Subject line

Subject: "Seu pedido #XXXXXX está pronto para retirada!" where `XXXXXX` is the first 8 characters of the checkout unique_id (UUID), styled in monospace. The full UUID is used in the email body; only the subject is truncated.

### New vs existing user detection

`clientmodel.Create` now returns `(clientID int, isNew bool, err error)`. `CreateAndAssociate` also returns `isNew`. The existing-user path inside `Create` (email already exists, reuse) sets `isNew = false`. Password hashing always happens before calling `Create` to keep the flow simple; one wasted bcrypt hash on the uncommon existing-user path is acceptable.

### Order data resolution

A new function `checkoutmodel.ResolvePickupEmailData` takes `[]ConfirmItem` (package IDs + quantities) and a database connection, runs a new SQL query joining packages → products → consignments → pickup_locations, and returns grouped `[]mail.PickupGroup`. The grouping collapses repeated product names within the same pickup location (summing quantities).

### Branding

The platform name "PVAuto" remains hardcoded in the template (unchanged from the current welcome template). The logo remains an embedded base64 PNG data URI.

### Login section format

For new users, the login section includes: user name, email, temporary password (in a styled info box), a "Configurar minha conta" button pointing to the login URL, and a secondary "Redefinir senha" link pointing to the reset URL. This matches the current welcome template structure; only the surrounding context changes.

### Spy mailer update

`SpyMailer.SentEmails` is renamed to `SpyMailer.SentConfirmation` of type `*PickupEmailData` (latest email only, matching the single-checkout-per-test pattern). The `SendWelcome` method is removed.

## Testing Decisions

### What makes a good test

Tests assert on external behavior: rendered template contains expected strings (product name, pickup location, order ID, QR code data URI, conditional presence/absence of login section), spy mailer records correct data, and integration tests exercise the full checkout-to-email pipeline.

### Modules to test

- **mail/mail_test.go**: Update template rendering tests. Add assertions for order info (product name, pickup location, order ID), QR code in output, and conditional login section (present when IsNewUser=true, absent when false). HTML escape test continues to apply.
- **api_checkout_client_test.go**: Update `SpyMailer` assertions to use `PickupEmailData`. Existing test `welcome_email_sent_on_client_creation` now validates that groups contain product names and pickup locations. The `existing_client_reused` test now additionally asserts that `IsNewUser` is false and no password is present.
- **ResolvePickupEmailData grouping**: Add an integration test (in `api_checkout_client_test.go`) that creates a checkout with items spanning multiple pickup locations and verifies the resolved groups are correctly partitioned.

### Prior art

- `mail/mail_test.go` — `TestRenderWelcome` / `TestRenderWelcomeHTMLEscaped` for template content assertion pattern
- `api_checkout_client_test.go` — `TestCheckoutClientAutoCreation` / `welcome_email_sent_on_client_creation` for spy mailer assertions in the checkout flow

## Out of Scope

- Activation/passwordless link flow (wireframe showed "request a secure activation link" — this is a separate auth refactor)
- Dynamic platform naming via environment variable
- Multi-language email templates
- Email delivery retry logic or queueing
- Changing the QR code format (remains base64 data URI PNG)
- Changing the password generation algorithm
- Sending emails for checkouts without client_email (unchanged behavior)

## Further Notes

- The `welcome.html` file is renamed to `pickup_confirmation.html`. The embedded Go variable and `init()` setup are updated accordingly.
- The old `Client.SendWelcome` method is fully removed — no backward compatibility layer.
- The `NoopMailer` implementation of the new method is a single no-op line (unchanged pattern).
- The order confirmation email is always sent best-effort after the HTTP response, same as the current welcome email. Failures are logged but never fail the checkout.

Triage: ready-for-agent
