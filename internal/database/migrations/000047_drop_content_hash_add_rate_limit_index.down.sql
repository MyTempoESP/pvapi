-- Restore content_hash dedup (pvapi-zle6 rollback)

BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN content_hash TEXT NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkouts_content_hash_unique UNIQUE (content_hash);

ALTER TABLE IF EXISTS public.checkouts
    ALTER COLUMN content_hash DROP DEFAULT;

DROP INDEX IF EXISTS idx_checkouts_client_processed_at;

COMMIT;
