BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    DROP CONSTRAINT IF EXISTS checkout_cart_ref_unique;

COMMIT;
