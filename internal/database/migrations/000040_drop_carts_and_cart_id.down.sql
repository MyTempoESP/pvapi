-- Restore cart subsystem tables and columns

BEGIN;

-- Recreate carts table
CREATE TABLE IF NOT EXISTS public.carts
(
    id serial NOT NULL,
    client_id integer,
    session_id text NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT cart_user_unique UNIQUE (client_id),
    CONSTRAINT cart_session_unique UNIQUE (session_id)
);

-- Recreate cart_product table
CREATE TABLE IF NOT EXISTS public.cart_product
(
    id serial NOT NULL,
    product_id integer NOT NULL,
    cart_id integer NOT NULL,
    personalizations jsonb NOT NULL,
    quantity integer NOT NULL,
    PRIMARY KEY (id)
);

-- Restore cart_id column on checkouts
ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN IF NOT EXISTS cart_id integer;

-- Restore FK constraint
ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkout_cart_relation FOREIGN KEY (cart_id)
    REFERENCES public.carts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

-- Restore index
CREATE INDEX IF NOT EXISTS fki_payments_cart
    ON public.checkouts(cart_id);

-- Restore cart_product FKs
ALTER TABLE IF EXISTS public.cart_product
    ADD CONSTRAINT product_cart_relation_through_cart_product FOREIGN KEY (product_id)
    REFERENCES public.products (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

ALTER TABLE IF EXISTS public.cart_product
    ADD CONSTRAINT cart_product_through_cart_product FOREIGN KEY (cart_id)
    REFERENCES public.carts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

-- Add price column to cart_product (added by 000015)
ALTER TABLE IF EXISTS public.cart_product
    ADD COLUMN IF NOT EXISTS price numeric(12, 2) NOT NULL DEFAULT 0;

COMMIT;
