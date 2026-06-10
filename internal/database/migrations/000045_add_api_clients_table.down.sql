-- Revert api_clients table and api_client_id on checkouts (pvapi-e9q3)

BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS api_client_id;

DROP TABLE IF EXISTS public.api_clients;

COMMIT;
