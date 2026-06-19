BEGIN;

ALTER TABLE public.packages
    ADD COLUMN status text NOT NULL DEFAULT 'active';

ALTER TABLE public.packages
    ADD CONSTRAINT packages_status_check CHECK (status IN ('active', 'paused'));

COMMIT;
