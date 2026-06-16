BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN IF NOT EXISTS callback text;

COMMIT;
