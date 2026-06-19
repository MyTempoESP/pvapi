BEGIN;

ALTER TABLE public.package_reserved
    ADD COLUMN personalizacao jsonb NOT NULL DEFAULT '{}'::jsonb;

COMMIT;
