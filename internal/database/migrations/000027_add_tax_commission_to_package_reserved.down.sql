BEGIN;

ALTER TABLE public.package_reserved
    DROP COLUMN tax,
    DROP COLUMN commission;

COMMIT;
