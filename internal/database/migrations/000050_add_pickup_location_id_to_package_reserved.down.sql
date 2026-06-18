BEGIN;

ALTER TABLE IF EXISTS public.package_reserved
    DROP COLUMN IF EXISTS pickup_location_id;

COMMIT;
