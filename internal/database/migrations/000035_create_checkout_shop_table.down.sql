BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS value_discrepancy;

DROP TABLE IF EXISTS public.checkout_shop;

COMMIT;
