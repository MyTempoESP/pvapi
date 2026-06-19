BEGIN;

ALTER TABLE public.packages
    DROP CONSTRAINT packages_status_check;

ALTER TABLE public.packages
    DROP COLUMN status;

COMMIT;
