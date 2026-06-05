-- Add CPF column to clients table (pvapi-yef3)

BEGIN;

ALTER TABLE public.clients ADD COLUMN IF NOT EXISTS cpf TEXT DEFAULT NULL;

COMMIT;
