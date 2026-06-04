---
# pvapi-xsfc
title: Delete cart subsystem
status: todo
type: task
priority: normal
created_at: 2026-06-04T23:05:25Z
updated_at: 2026-06-04T23:05:31Z
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
