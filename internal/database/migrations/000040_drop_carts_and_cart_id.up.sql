-- Drop cart subsystem tables and columns
-- Checkout no longer depends on carts (pvapi-xsfc)

BEGIN;

-- Drop FK constraint from checkouts to carts
ALTER TABLE IF EXISTS public.checkouts
    DROP CONSTRAINT IF EXISTS checkout_cart_relation;

-- Drop index on checkouts.cart_id
DROP INDEX IF EXISTS fki_payments_cart;

-- Drop cart_id column from checkouts
ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS cart_id;

-- Drop cart_product table first (FK to carts)
DROP TABLE IF EXISTS public.cart_product CASCADE;

-- Drop carts table
DROP TABLE IF EXISTS public.carts CASCADE;

COMMIT;
