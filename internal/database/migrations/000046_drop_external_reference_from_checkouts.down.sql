-- Restore external_reference column (pvapi-37pg rollback)

BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN external_reference TEXT NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS public.checkouts
    ALTER COLUMN external_reference DROP DEFAULT;

COMMIT;
