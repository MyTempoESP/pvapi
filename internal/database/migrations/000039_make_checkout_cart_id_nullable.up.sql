-- Make cart_id nullable in checkouts since checkout no longer requires a cart
ALTER TABLE IF EXISTS public.checkouts
    ALTER COLUMN cart_id DROP NOT NULL;
