BEGIN;

ALTER TABLE IF EXISTS public.checkout_shop
    DROP CONSTRAINT IF EXISTS checkout_shop_client_relation;

ALTER TABLE IF EXISTS public.checkout_shop
    DROP COLUMN IF EXISTS client_id;

COMMIT;
