-- Remove unique_id from package_reserved (pvapi-lgua)

BEGIN;

ALTER TABLE IF EXISTS public.package_reserved
    DROP COLUMN IF EXISTS unique_id;

COMMIT;
