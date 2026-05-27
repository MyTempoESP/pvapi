---
# pvapi-3jmd
title: Change pix_cpf to generic pix_key field
status: completed
type: task
priority: normal
created_at: 2026-05-26T14:01:51Z
updated_at: 2026-05-26T15:47:44Z
---

Triage: ready-for-agent

## Problem Statement

When requesting a withdrawal (saque), shop owners must provide a PIX key for the destination account. The current API only accepts a CPF as a PIX key via the `pix_cpf` field. Brazilian PIX keys come in multiple formats — email, phone, random UUID, CNPJ — none of which are currently supported. Shop owners with non-CPF PIX keys cannot request withdrawals.

## Solution

Rename the `pix_cpf` field to `pix_key` across the entire stack, removing the implied CPF constraint. The field remains a free-text string with no format validation. API consumers send any valid PIX key type; the downstream banking provider handles validation and returns specific error messages for invalid keys.

## User Stories

1. As a shop owner with an email-based PIX key, I want to submit my email as a PIX key when requesting a withdrawal, so that I can receive funds to my registered PIX key.
2. As a shop owner with a phone-based PIX key, I want to submit my phone number as a PIX key, so that I can receive funds without needing a CPF-based key.
3. As a shop owner with a random (UUID) PIX key, I want to submit that key, so that I can use the key format I registered with my bank.
4. As a shop owner with a CNPJ-based PIX key (business account), I want to submit my CNPJ, so that withdrawals go to my business account.
5. As an API consumer, I want the request and response field names to match, so that I don't need to remember different naming conventions for input vs output.
6. As a developer integrating with the withdrawals API, I want the field name `pix_key` to self-document its purpose, so that I don't mistakenly assume only CPF is accepted.

## Implementation Decisions

- **Database migration**: Create a new migration (`ALTER TABLE saques RENAME COLUMN pix_cpf TO pix_key`). Do not edit the existing 000028 migration.
- **Clean break**: No backwards compatibility. The `pix_cpf` JSON field, Go struct field, and DB column are all renamed to `pix_key` in a single deploy. The saque endpoint has zero production traffic.
- **No validation**: The field remains a free-text `*string`. No server-side format validation is added. Invalid PIX keys are rejected downstream by the banking provider.
- **Response consistency**: The nested response field `pix.cpf` is also renamed to `pix.key`. API responses will show `{"pix": {"key": "...", "titular": "..."}}`.
- **`pix_titular` unchanged**: The account holder name field is not renamed. It remains useful for CPF/CNPJ key types.
- **Modules modified**: The saque request/response models, the database row struct, SQL queries, and the API fixture. No new modules are created — this is a rename, not new functionality.

## Testing Decisions

- Update the existing fixture in `api_fixtures_test.go` to use `"pix_key"` instead of `"pix_cpf"`.
- No new tests are needed. A non-CPF PIX key (email, phone) is just a string — Go and PostgreSQL don't distinguish formats.
- Run the full test suite with the test-code skill to confirm nothing is broken.

## Out of Scope

- PIX key format validation (server-side).
- Renaming `pix_titular`.
- Adding support for multiple PIX keys per withdrawal.
- Any changes to the downstream banking integration.

## Further Notes

- This is a breaking change. No clients currently integrate against this endpoint, so no deprecation period is needed.

## Summary of Changes

- **models/saquemodel/requests.go**: Renamed `PixCPF` field to `PixKey` with JSON tag `pix_key`
- **models/saquemodel/model.go**: Renamed `PixCPF` struct field and `db:"pix_cpf"` tag to `PixKey` / `db:"pix_key"`; updated constructor parameter
- **models/saquemodel/resources.go**: Renamed `PixResource.CPF` to `PixResource.Key` with JSON tag `key`; updated all references from `r.PixCPF` / `sr.Pix.CPF` to `r.PixKey` / `sr.Pix.Key`
- **query/saque.go**: Replaced all 3 SQL occurrences of `pix_cpf` with `pix_key`
- **api_fixtures_test.go**: Updated fixture from `"pix_cpf"` to `"pix_key"`
- **scripts/test-saques.sh**: Updated JSON body from `pix_cpf` to `pix_key`
- **internal/database/migrations/000030**: New migration with `ALTER TABLE saques RENAME COLUMN pix_cpf TO pix_key` (+ down migration)

All 99 tests pass. Lint clean.
