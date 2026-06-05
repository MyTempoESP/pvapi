-- Remove CPF column from clients table

BEGIN;

ALTER TABLE public.clients DROP COLUMN IF EXISTS cpf;

COMMIT;
