---
# pvapi-xsfc
title: Delete cart subsystem
status: completed
type: task
priority: normal
created_at: 2026-06-04T23:05:25Z
updated_at: 2026-06-05T00:56:19Z
blocked_by:
    - pvapi-i5ju
---

## What to build

Delete all cart-related code now that checkout no longer depends on it.

Changes:
- Drop `carts` and `cart_product` tables (down migration)
- Drop `checkouts.cart_id` column and its FK constraint
- Delete `controllers/cart/cart.go`
- Delete `models/cartmodel/` (model.go, requests.go, resources.go, blueprint.go)
- Delete `query/cart.go`
- Remove `/carrinho/*` route group from routes.go
- Remove `cartmodel` import from any remaining files

## Acceptance criteria

- [ ] `carts` and `cart_product` tables no longer exist
- [ ] `checkouts.cart_id` column no longer exists
- [ ] No `/carrinho` routes respond
- [ ] `controllers/cart/`, `models/cartmodel/`, `query/cart.go` are deleted
- [ ] Full test suite passes
- [ ] `go build` succeeds

## Blocked by

- pvapi-i5ju



## Summary of Changes

All cart-related code has been removed:

### Migration
- Created  — drops  and  tables, drops  column and its FK/index
- Created  — full restoration of cart subsystem

### Deleted files
- 
-  (model.go, requests.go, resources.go, blueprint.go)
- 
- 

### Edited files
-  — removed cart import and  route group
-  — removed , , 
-  — removed dead  and  functions
-  — removed dead  function
-  — removed  helper and its call in 

### Verification
-  — passes
-  — passes
- No remaining cart references in  files

## Summary of Changes

All cart-related code has been removed:

### Migration
- Created 000040_drop_carts_and_cart_id.up.sql — drops carts and cart_product tables, drops checkouts.cart_id column and its FK/index
- Created 000040_drop_carts_and_cart_id.down.sql — full restoration of cart subsystem

### Deleted files
- controllers/cart/cart.go
- models/cartmodel/ (model.go, requests.go, resources.go, blueprint.go)
- query/cart.go
- api_cart_checkout_test.go

### Edited files
- routes.go — removed cart import and /carrinho/* route group
- query/client.go — removed cart-related queries
- models/clientmodel/model.go — removed dead AssociateToCart() and SetCartClientID() functions
- models/productmodel/model.go — removed dead TakeFromCart() function
- api_fixtures_test.go — removed createCartItem helper and its call in generateEarnings

### Verification
- go build ./... passes
- go vet ./... passes
- No remaining cart references in .go files
