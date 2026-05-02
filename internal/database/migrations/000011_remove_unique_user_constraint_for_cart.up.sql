-- Allow users to have multiple carts (especially FINISHED_ ones)
ALTER TABLE IF EXISTS public.carts DROP CONSTRAINT cart_user_unique;