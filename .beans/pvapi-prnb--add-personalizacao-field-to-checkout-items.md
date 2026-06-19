---
# pvapi-prnb
title: Add personalizacao field to checkout items
status: todo
type: feature
priority: normal
created_at: 2026-06-19T04:15:40Z
updated_at: 2026-06-19T04:15:42Z
---

## Problem Statement

External integrations sending checkouts through `/checkouts/confirm` have no way to attach per-item customization metadata (e.g., `"tamanho": "G"`, `"cor": "azul"`) that flows through the system and back in the response. This metadata is needed for fulfillment, labeling, and consumer-facing order receipts.

## Solution

Add a `personalizacao` field to each item in the `/checkouts/confirm` request and response. The field carries arbitrary string key-value pairs, is optional, and defaults to an empty object (`{}`). It is stored per line-item in `package_reserved` alongside the existing financial breakdown.

## User Stories

1. As an API client, I want to include custom key-value metadata per checkout item, so that fulfillment systems can see variant selections (size, color, engraving text).
2. As an API client, I want the personalizacao I sent to be echoed back in the confirm response, so that I can reconcile my request with the platform's response.
3. As an API client, I want an empty `personalizacao` to be valid and silent, so that existing integrations without this field continue working without changes.
4. As a shop owner, I want to see the personalizacao on items in future checkout detail views, so that I know what specific variant the customer selected.
5. As an operator, I want the personalizacao data stored per reserved package row, so that when a checkout is partially canceled the remaining items still carry their customization.
6. As a developer, I want personalizacao to be a flat `map[string]string`, so that querying and display are simple and predictable.

## Implementation Decisions

- **New database column:** `package_reserved.personalizacao jsonb NOT NULL DEFAULT '{}'::jsonb`. Migration `000055`.
- **Value type:** `map[string]string` — both keys and values are always strings. No nesting, no numbers, no booleans.
- **Optional field:** empty `personalizacao: {}` is valid and is the default. The field is always present in responses.
- **Request DTO (`ConfirmItem`):** add `Personalizacao map[string]string json:"personalizacao"`.
- **Response DTO (`QRCodeEntry`):** add `Personalizacao map[string]string json:"personalizacao"`.
- **Insert query (`InsertPackageReservedQuery`):** accept and insert the new column using named parameter `@personalizacao`.
- **Enriched select query (`EnrichedQRCodeEntriesQuery`):** select `pr.personalizacao` and map it to the response DTO.
- **No validation:** no limit on number of keys, no key name restrictions, no size limits. The jsonb column enforces valid JSON and that's sufficient.
- **Not propagated** to cancel response, email templates, admin listing, or shop checkout endpoints at this time (see Out of Scope).

## Testing Decisions

- **What makes a good test:** test through the HTTP layer with the API harness, asserting the request/response contract. Do not test internal model functions in isolation — they are covered transitively.
- **Prior art:** `api_checkout_client_test.go` — `TestCheckoutClientAutoCreation` and related subtests that exercise `/checkouts/confirm` end-to-end with `apiHarness`.
- **Tests to add:**
  - Confirm with `personalizacao` on items → response `qr_codes[].personalizacao` matches.
  - Confirm without `personalizacao` → response `qr_codes[].personalizacao` is `{}`.
  - Confirm with empty `personalizacao: {}` → response mirrors it.
  - Confirm with multiple items, each with different personalizacao → each response entry carries its own.

## Out of Scope

- Displaying personalizacao in cancel response (`/checkouts/:unique_id/cancel`).
- Displaying personalizacao in shop checkout listing (`/checkouts` for shop users).
- Displaying personalizacao in admin checkout views.
- Personalizacao in email templates (reservation confirmation, pickup notification).
- Personalizacao on GET `/checkouts/:unique_id` (if such endpoint exists or is added later).

## Further Notes

- The `checkouts` table already has a `metadata jsonb` column. This PRD does NOT use it — the data is per-line-item, not per-checkout.
- Future: if consumers want to query by personalizacao keys, a GIN index on `package_reserved.personalizacao` can be added. Not needed now.

Triage: ready-for-agent
