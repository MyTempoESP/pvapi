---
# pvapi-w1n8
title: Remove cart-checkout skill
status: completed
type: task
priority: normal
created_at: 2026-06-04T23:05:38Z
updated_at: 2026-06-05T01:01:01Z
blocked_by:
    - pvapi-xsfc
---

## What to build

Delete the cart-checkout skill since carts no longer exist and the checkout flow no longer uses them.

Changes:
- Delete `.pi/skills/cart-checkout/SKILL.md`
- Delete `.pi/skills/cart-checkout/scripts/checkout.sh`
- Remove the `cart-checkout` entry from available skills if centrally registered

## Acceptance criteria

- [ ] `.pi/skills/cart-checkout/` directory no longer exists
- [ ] No references to cart-checkout skill remain in project configuration

## Blocked by

- pvapi-xsfc



## Summary of Changes

- Deleted `.pi/skills/cart-checkout/SKILL.md`
- Deleted `.pi/skills/cart-checkout/scripts/checkout.sh`
- Removed `cart-checkout` row from the API reference table in `.pi/skills/api-request/SKILL.md`
- No other project references to cart-checkout remain
