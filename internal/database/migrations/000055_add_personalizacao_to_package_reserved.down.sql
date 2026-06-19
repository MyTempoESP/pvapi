BEGIN;

ALTER TABLE public.package_reserved
    DROP COLUMN personalizacao;

COMMIT;
