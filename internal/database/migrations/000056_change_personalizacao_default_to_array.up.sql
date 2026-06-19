BEGIN;

ALTER TABLE public.package_reserved
    ALTER COLUMN personalizacao SET DEFAULT '[]'::jsonb;

COMMIT;
