---
# pvapi-yje0
title: 'Bank Accounts: reusable bank account object for shops'
status: todo
type: feature
priority: normal
created_at: 2026-05-27T01:50:25Z
updated_at: 2026-05-27T01:50:49Z
---

## Problem Statement

Shops (establishments) that create saques (withdrawal requests) must manually re-enter bank details every time they submit a withdrawal. There is no way to save a bank account for reuse. This is repetitive and error-prone.

## Solution

Introduce a "bank account" resource that shops can create, list, update, and delete. Saque creation will reference a saved bank account instead of accepting raw bank details inline.

## User Stories

1. As a shop, I want to create a saved bank account with a name, bank name, PIX key/titular, and optional transfer details, so that I can reuse it across withdrawal requests.
2. As a shop, I want to list all my saved bank accounts, so that I can see what accounts I have available.
3. As a shop, I want to update a saved bank account (name, bank, PIX info, transfer details), so that I can correct mistakes or reflect changes.
4. As a shop, I want to delete a saved bank account I no longer use, so that my account list stays clean.
5. As a shop, I want to create a saque by selecting one of my saved bank accounts instead of typing all bank details again, so that withdrawal is faster and less error-prone.
6. As a shop, I want to see the same bank information in saque responses (bank name, PIX, transfer details) regardless of whether it came from a saved account, so that existing consumers of the API are unaffected.
7. As a developer, I want bank account ownership enforced so that shops cannot view, modify, or delete another shop bank account.

## Implementation Decisions

- **Routes**: bank account CRUD lives under `/estabelecimentos/contasbancarias`. POST (create) and GET (list) use role check; PUT and DELETE use ownership check via HasBankAccount. All use path params for identifying individual accounts (e.g. `DELETE /estabelecimentos/contasbancarias/:id`).
- **No individual view endpoint**: only list (GET) and create/update/delete.
- **Saque creation contract**: the `POST /estabelecimentos/saques` body no longer accepts `bank_name`, `pix_key`, `pix_titular`, or `transfer_details`. Instead it accepts a `bank_account` field containing the bank account ID. The model layer validates the account exists and belongs to the requesting shop.
- **Saque responses unchanged**: saque resources still include `bank_name`, `pix`, and `transfer_details` — sourced via JOIN on bank_accounts rather than from saque columns.
- **Bank account resource**: uses a nested `pix: { key, titular }` sub-object (consistent with saques PixResource), includes `shop_id` as read-only, `transfer_details` as optional JSON, `created_at` as RFC3339 string.
- **Partial update semantics**: PUT on bank accounts accepts only the fields to change (nil = unchanged). `shop_id` and `created_at` are not writable.
- **Nullable PIX**: `pix_key` and `pix_titular` are nullable — a shop may create a bank account for TED/DOC-only transfers without PIX info.
- **Name required**: the `name` field on bank accounts is required (user-facing label like "Conta Principal", "Nubank PF").
- **Deletion**: deleting a bank account sets the FK in saques to NULL (ON DELETE SET NULL). No restriction on deletion beyond ownership check.
- **Listing**: flat array, no pagination.
- **Authorization**: HasBankAccount lives in shopmodel alongside other Has* functions. Gate follows the existing Saques dispatch pattern.
- **Schema**:
  - New `bank_accounts` table: id SERIAL, shop_id FK→shops, name TEXT NOT NULL, bank_name TEXT, pix_key TEXT (nullable), pix_titular TEXT (nullable), transfer_details JSONB, created_at TIMESTAMPTZ.
  - Altered `saques` table: drop pix_key, pix_titular, transfer_details; add bank_account_id INT NOT NULL FK→bank_accounts. bank_name column dropped. Existing saque rows will be lost (no production data).
  - Two migrations: 000031 (create bank_accounts), 000032 (alter saques).

## Testing Decisions

- **What makes a good test**: test external API behavior only. Each test/subtest has a single well-defined objective. Repeated setup blocks extracted to fixture helpers.
- **Modules tested**: integration tests via HTTP (`api_bank_account_test.go`); modified saque tests in existing file.
- **Prior art**: tests follow the `TestSaque` pattern — single top-level test function with descriptive subtests (`create_missing_fields`, `create_unauthorized`, `list`, `update`, `delete`, etc.). Fixture helpers in `api_fixtures_test.go` provide `createBankAccount` and updated `createSaque`.
- **Test coverage**: all CRUD operations, authorization (unauthorized access attempts), validation (missing fields, invalid data), cross-shop isolation.

## Out of Scope

- Individual bank account view endpoint (GET by ID)
- Updating a saque bank account reference (UpdateSaqueRequest unchanged)
- Bulk operations on bank accounts
- Pagination for bank account listing
- Bank account usage count / metrics

## Further Notes

- All terminology follows the domain glossary: "shop" = establishment, "saque" = withdrawal request.
- The `name` field on bank_accounts must be required on creation.
- Bank account ID in saque creation body is validated: model layer checks existence and shop ownership before insert.
