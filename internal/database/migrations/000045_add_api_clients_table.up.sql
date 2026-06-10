-- Add api_clients table and api_client_id to checkouts (pvapi-e9q3)

BEGIN;

CREATE TABLE IF NOT EXISTS public.api_clients (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN IF NOT EXISTS api_client_id INTEGER REFERENCES public.api_clients(id) ON DELETE SET NULL;

COMMIT;
