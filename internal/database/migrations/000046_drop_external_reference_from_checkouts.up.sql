-- Drop external_reference from checkouts (pvapi-37pg)

BEGIN;

ALTER TABLE IF EXISTS public.checkouts DROP COLUMN IF EXISTS external_reference;

COMMIT;
