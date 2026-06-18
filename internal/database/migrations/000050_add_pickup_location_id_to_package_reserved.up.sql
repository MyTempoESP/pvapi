BEGIN;

ALTER TABLE IF EXISTS public.package_reserved
    ADD COLUMN IF NOT EXISTS pickup_location_id INTEGER NOT NULL REFERENCES public.pickup_locations(id);

COMMIT;
