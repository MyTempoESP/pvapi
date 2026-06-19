BEGIN;

ALTER TABLE public.pickup_locations
    RENAME COLUMN pickup_at TO starts_at;

ALTER TABLE public.pickup_locations
    ADD COLUMN description TEXT NOT NULL DEFAULT '';

ALTER TABLE public.pickup_locations
    ADD COLUMN valid_until TIMESTAMPTZ NOT NULL DEFAULT '2099-12-31 23:59:59+00';

ALTER TABLE public.pickup_locations
    ALTER COLUMN description DROP DEFAULT,
    ALTER COLUMN valid_until DROP DEFAULT;

COMMIT;
