BEGIN;

ALTER TABLE public.package_reserved
    ADD COLUMN tax numeric(12, 2) NOT NULL DEFAULT 0,
    ADD COLUMN commission numeric(12, 2) NOT NULL DEFAULT 0;

COMMIT;
