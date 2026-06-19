BEGIN;

ALTER TABLE public.pickup_locations
    DROP COLUMN IF EXISTS valid_until;

ALTER TABLE public.pickup_locations
    DROP COLUMN IF EXISTS description;

ALTER TABLE public.pickup_locations
    RENAME COLUMN starts_at TO pickup_at;

COMMIT;
