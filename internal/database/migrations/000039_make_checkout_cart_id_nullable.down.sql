-- Restore NOT NULL on cart_id (fails if any null rows exist)
ALTER TABLE IF EXISTS public.checkouts
    ALTER COLUMN cart_id SET NOT NULL;
