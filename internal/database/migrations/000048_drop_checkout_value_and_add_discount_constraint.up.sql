-- Drop value and value_discrepancy from checkouts (no longer needed;
-- financial truth lives in checkout_shop.shop_value).
-- Also add a check constraint on packages.discount.

BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS value;

ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS value_discrepancy;

ALTER TABLE IF EXISTS public.packages
    ADD CONSTRAINT discount_range CHECK (discount >= 0 AND discount <= 100);

COMMIT;
