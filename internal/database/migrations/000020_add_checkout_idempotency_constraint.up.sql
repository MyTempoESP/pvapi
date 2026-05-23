BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkout_cart_ref_unique UNIQUE (cart_id, external_reference);

COMMIT;
