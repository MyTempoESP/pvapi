---
# pvapi-a4cn
title: Create a new "bank account" object for user experience
status: completed
type: task
priority: normal
created_at: 2026-05-26T14:07:39Z
updated_at: 2026-05-27T01:56:54Z
parent: pvapi-yje0
---

## Terminology

- mentions to `user` in this document refer strictly to a shop, identified by
a JWT token.

## Behaviour

There is currently no way for a user to "save" a bank account used in a `saque`.

## Expected / Requirements

The user must be able to re-utilize bank accounts. The following must be in place:

- A user may create new bank accounts manually through POST /estabelecimentos/contasbancarias
- A user may list bank accounts through GET /estabelecimentos/contasbancarias, the list includes each full bank account resource, containing all fields
- A user may delete bank accounts through DELETE /estabelecimentos/contasbancarias, without restrictions other than being logged-in
- A user may update bank accounts through PUT /estabelecimentos/contasbancarias, all fields editable, except for shop_id and created_at, which do not make sense to update
- The `saque` creation must be changed to allow a bank_account field instead

## Implementation details

- New database migration for bank accounts, containing the following fields:
	- id SERIAL
	- shop_id FK -> shops
	- name TEXT
	- bank_name TEXT
	- pix_key TEXT
	- pix_titular TEXT
	- transfer_details JSONB
	- created_at (codebase standard)
- New database migration (after bank accounts) editing `saques`, with the following:
	- Drop the `pix_key`, `pix_titular`, `transfer_details` fields
	- Add `bank_account_id` FK -> bank_accounts
- POST /estabelecimentos/saques should not receive the bank data directly, instead receiving the usual fields and a "bank_account" containing the bank account id
- Create the necessary `contasbancarias` handlers in a dedicated bank account controller under controllers/bankaccount/bankaccount.go
- Create the associated `contasbancarias` operations and database interactions in models/bankaccountmodel/model.go
- Create the request objects and validations under models/bankaccountmodel/requests.go
- Create the return resources and formatting under models/bankaccountmodel/resources.go
- Add the route associations in ./routes.go
- Edit gate/shop.go and make sure to allow access for the shop to its own bank accounts
	- For update/deletion, there must be a `shop.HasBankAccount` function to assure ownership
	- For creation/listing, a simple role check is enough

## Verification and Testing

The task will be considered complete after the following:

- All the currently implemented tests successfully pass
- There is a new api_bank_account_test.go file
- The `saques` tests are edited to include the bank account change
- The api_bank_account_test.go contains integration tests for all the requirements provided
- The individual api_bank_account_test.go tests pass a test quality check, meeting the following:
	- Repeated blocks must be extracted to helper functions
	- Each test / subtest has a well-defined and clear objective, and test ONLY that objective
- The individual api_bank_account_test.go tests successfully pass
- The individual saques tests successfully pass

## Summary of Changes

- Created migration 000031 (bank_accounts table) and 000032 (alter saques to use bank_account_id FK)
- Created models/bankaccountmodel/ (model.go, requests.go, resources.go) with CRUD operations
- Created controllers/bankaccount/bankaccount.go with Create, List, Update, Delete handlers
- Created query/bankaccount.go with all SQL queries including ShopHasBankAccountQuery
- Updated models/shopmodel/model.go with HasBankAccount function
- Updated gate/shop.go with BankAccounts dispatch rule (role check for shop role)
- Updated routes.go with /contasbancarias routes (POST, GET, PUT, DELETE)
- Updated models/saquemodel/ (model.go, requests.go, resources.go) to use bank_account_id instead of inline bank fields
- Updated query/saque.go to JOIN bank_accounts for bank info
- Updated controllers/saque/saque.go to validate bank_account ownership
- Updated api_fixtures_test.go with createBankAccount helper, updated createSaque signature
- Updated api_saque_test.go to use bank accounts
- Created api_bank_account_test.go with comprehensive integration tests
