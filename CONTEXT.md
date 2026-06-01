# PVAPI

E-commerce marketplace backend. Shops list products sourced from consignments; clients browse and purchase through carts; checkouts produce per-shop financial records.

## Language

**Shop (Estabelecimento)**:
A vendor that sells products through the platform.
_Avoid_: Store, seller, merchant

**Client (Cliente)**:
A person who purchases products from shops.
_Avoid_: Customer, buyer, user

**Consignee (Monitor)**:
An operator responsible for receiving and managing consignments at pickup locations.
_Avoid_: Monitor, receiver

**Consignment (Remessa)**:
A batch of products shipped by a shop to a consignment location.
_Avoid_: Shipment, delivery, batch

**Checkout**:
A confirmed purchase transaction. A single checkout may span products from multiple shops, producing one `checkout_shop` row per shop involved.
_Avoid_: Order, purchase, transaction

**Cart (Carrinho)**:
A session-bound collection of products before checkout. Each cart is tied to a session ID and optionally to a client.
_Avoid_: Basket, bag

**Pickup Location (Ponto)**:
A physical location where clients collect purchased products.
_Avoid_: Pickup point, collection point

**Package (Pacote)**:
A priced unit of a product at a specific pickup location, derived from a consignment. Defines price, tax, and commission.
_Avoid_: SKU, listing, variant

**Saque (Withdrawal)**:
A shop's request to withdraw earnings from the platform to a bank account.
_Avoid_: Withdrawal request, payout

## Example dialogue

> **Dev:** When a client checks out with products from two different shops, how many checkout rows are created?
>
> **Domain expert:** One row in `checkouts` — it's a single transaction from the client's perspective. But two rows in `checkout_shop`, one per shop, each with that shop's share of the value, tax, and commission.
>
> **Dev:** And the client is linked to both `checkout_shop` rows?
>
> **Domain expert:** Yes, `checkout_shop.client_id` points to the same client in both rows. That's how we can answer "how much has this client purchased from this shop?"
>
> **Dev:** What if the checkout is anonymous — no client email provided?
>
> **Domain expert:** Then `checkout_shop.client_id` stays NULL. Those purchases aren't attributed to any client.
