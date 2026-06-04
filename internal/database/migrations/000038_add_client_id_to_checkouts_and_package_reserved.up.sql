BEGIN;

-- Add client_id to checkouts with FK to clients(id)
ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN IF NOT EXISTS client_id integer;

ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkouts_client_relation FOREIGN KEY (client_id)
    REFERENCES public.clients (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

-- Add UNIQUE(external_reference) for idempotency
ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkouts_external_reference_unique UNIQUE (external_reference);

-- Add client_id to package_reserved with FK to clients(id)
ALTER TABLE IF EXISTS public.package_reserved
    ADD COLUMN IF NOT EXISTS client_id integer;

ALTER TABLE IF EXISTS public.package_reserved
    ADD CONSTRAINT package_reserved_client_relation FOREIGN KEY (client_id)
    REFERENCES public.clients (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

COMMIT;
