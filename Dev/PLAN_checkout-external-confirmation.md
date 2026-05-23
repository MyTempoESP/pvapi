# PLAN: External Payment Checkout Confirmation

## Overview

Clients pay on an external site. That external system calls our API to confirm payment,
reserve stock, and issue a QR code (the checkout `unique_id`). Later, a consignee scans
the QR code and the checkout transitions from `paid` → `received`.

Stock is tracked via a new `package_reserved` table. Packages are never mutated directly;
available stock = `packages.quantity - SUM(package_reserved.quantity)`.

---

## Current State

### checkouts table (migration 000009)
```
id                 serial         PK
cart_id            integer        FK → carts
consignment_id     integer        nullable FK → consignments
unique_id          uuid           DEFAULT uuidv4()
value              numeric(12,2)  NOT NULL
status             text           NOT NULL
description        text           NOT NULL
external_reference text           NOT NULL
metadata           jsonb          nullable
processed_at       timestamptz    DEFAULT now()
```

### packages table (migration 000001)
```
id               serial         PK
product_id       integer        FK → products
consignment_id   integer        FK → consignments
quantity         integer
price            numeric(12,2)
discount         numeric(12,2)
featured         boolean
attributesvalues jsonb
```

### carts table constraints
- `cart_user_unique` — **dropped** (migration 000011)
- `cart_session_unique` — **still active** → must be replaced with conditional index

### Dead code
- `CartHoldStock` / `product_decrement_inventory` / `TakeFromPackage` are no longer used
  for payment flow (no in-house payments). Leave them; they may still serve consignment
  management.

---

## Requirements

### R1 — External payment confirmation endpoint
`POST /checkouts/confirm`

**Request:**
```json
{
  "cart_id":            123,
  "value_paid":         "150.00",
  "external_reference": "MP-ABC123",
  "items": [
    {"product_id": 1, "consignment_id": 3, "quantity": 5},
    {"product_id": 2, "consignment_id": 3, "quantity": 2},
    {"product_id": 3, "consignment_id": 7, "quantity": 1}
  ]
}
```

- `items` is the "consignment reference" — each entry maps a product to its consignment.
- A single checkout may span 2–3 consignments (carts cross shops).

**Behavior (single atomic transaction):**
1. Idempotency check: if `(cart_id, external_reference)` already exists → return existing `unique_id`
2. Validate cart exists
3. For each item, resolve `packages` row via `(product_id, consignment_id)`
4. Check availability: `packages.quantity - SUM(package_reserved.quantity)` ≥ requested
5. INSERT into `checkouts` with `status = 'paid'`, generated `unique_id`
6. INSERT into `package_reserved` (one row per item)
7. Soft-delete cart: `UPDATE carts SET checked_out_at = NOW()`
8. Return `{ "unique_id": "<uuid>" }`

### R2 — Stock reservation via package_reserved
- `packages.quantity` is **never decremented** by checkout flow
- `package_reserved` accumulates reservations per package
- Available = `packages.quantity - SUM(package_reserved.quantity)`
- Race condition protection: `SELECT ... FOR UPDATE` on packages rows

### R3 — Soft-delete carts after checkout
- After confirmation, overwrite `carts.session_id` with `FINISHED-<external_reference>`
- The `cart_session_unique` constraint is preserved — the original session_id is freed
  and a new cart can be created with it, while the finished cart retains a unique
  session_id for audit purposes
- No schema changes needed (no new column, no constraint drop)

### R4 — Checkout status state machine
```
paid ──────► received
```
- `paid`: payment confirmed, stock reserved, QR code issued
- `received`: consignee scanned QR code, product handed over

### R5 — Status transition endpoint (future, not in this implementation)
`POST /checkouts/:unique_id/receive`

- Out of scope for this phase. The endpoint will be implemented later.
- When built: validates `unique_id` exists, validates current status is `paid` (409 if not),
  updates to `received`, no stock mutation needed

### R6 — Idempotency
- Unique constraint on `checkouts(cart_id, external_reference)`
- Retries return the same `unique_id`

---

## Design Decisions

| Decision | Rationale |
|---|---|
| `package_reserved` instead of decrementing `packages.quantity` | Immutable audit trail; original allocation never lost; easy rollback |
| `consignment_id` on checkouts stays nullable | Single FK can't represent multi-consignment checkouts; relationship derived via `package_reserved → packages → consignments` |
| Overwrite `session_id` with `FINISHED-<ext_ref>` | Simpler than conditional index; no constraint changes; original session_id freed for reuse |
| `SELECT ... FOR UPDATE` on packages during availability check | Serializes concurrent confirmations on same consignment+product |
| One checkout row per confirmation | Simplifies idempotency; items live in `package_reserved` |

---

## Schema Changes

### Migration 000019 — package_reserved table
```sql
CREATE TABLE public.package_reserved (
    id               serial         PRIMARY KEY,
    package_id       integer        NOT NULL REFERENCES public.packages(id),
    checkout_id      integer        NOT NULL REFERENCES public.checkouts(id),
    quantity         integer        NOT NULL CHECK (quantity > 0),
    reserved_at      timestamptz    DEFAULT now()
);

CREATE INDEX idx_package_reserved_package ON public.package_reserved(package_id);
CREATE INDEX idx_package_reserved_checkout ON public.package_reserved(checkout_id);
```

### Migration 000020 — Idempotency constraint
```sql
ALTER TABLE public.checkouts
  ADD CONSTRAINT checkout_cart_ref_unique UNIQUE (cart_id, external_reference);
```

---

## Implementation Tasks (ordered)

### Phase 1 — Database
1. Write migration 000019 (package_reserved table)
2. Write migration 000020 (checkouts unique constraint)

### Phase 2 — Query layer (`query/checkout.go`, new file)
4. `InsertCheckout` — insert a checkout row, return `unique_id`
5. `InsertPackageReserved` — batch insert into package_reserved
6. `FindCheckoutByCartRef` — idempotency lookup by `(cart_id, external_reference)`
7. `CheckPackageAvailability` — FOR UPDATE availability check returning package_id + available
8. `UpdateCheckoutStatus` — transition status by unique_id
9. `ArchiveCart` — set checked_out_at on cart

### Phase 3 — Model layer (`models/checkoutmodel/`, new package)
10. `Checkout` struct (db-tagged for pgx scanning)
11. `ConfirmRequest` struct with validation tags
12. `Confirm()` — orchestrates the full transaction (validate → reserve → insert → archive)
13. `Receive()` — status transition with state validation

### Phase 4 — Controller (`controllers/checkout/checkout.go`, new file)
13. `POST /checkouts/confirm` handler — parse request, call `Confirm()`, return `unique_id`

### Phase 5 — Routes (`routes.go`)
14. Register `POST /checkouts/confirm`

### Phase 6 — Verification
15. Write tests in `api_requests_test.go`
    - Confirm checkout → 200 with unique_id
    - Idempotent retry → 200 with same unique_id
    - Invalid cart → 404
    - Empty cart → 422
    - Insufficient stock → 409
    - Product not in consignment → 422
16. `golangci-lint run --fix`
17. Manual test with real request

---

## Risks & Dependencies

| Risk | Mitigation |
|---|---|
| Cart with no items | Validate cart has `cart_product` rows before proceeding |
| Package not found for (product, consignment) | Return clear error: "product X not available in consignment Y" |
| Concurrent confirmations on same cart | `(cart_id, external_reference)` unique constraint prevents duplicates; FOR UPDATE serializes different references |
| `consignment_id` is nullable on checkouts | Acceptable — relationship is via package_reserved chain; can be formalized later with a junction table if needed |

---

## Open Questions

1. **Should we validate requested quantity against cart quantity?** If the external system sends `quantity: 10` but the cart only has `quantity: 3`, do we reject or trust the external system? Recommended: reject.

2. **What happens to cart items after checkout?** The cart's `session_id` is overwritten to `FINISHED-<ext_ref>`. The `cart_product` rows remain for audit. The original session_id is freed for a new cart.

3. **Multiple checkouts for same cart with different external_references?** The unique constraint on `(cart_id, external_reference)` allows this. If it shouldn't, add a constraint that only one non-cancelled checkout exists per cart. For MVP, not needed.

4. **Cancellation/refund flow?** Not in scope. When needed: DELETE from `package_reserved` by checkout_id, set checkout status to `cancelled`.
