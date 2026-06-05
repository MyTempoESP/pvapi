-- Add content_hash for content-based idempotency (pvapi-0bi4)

BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN content_hash TEXT NOT NULL DEFAULT '';

ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkouts_content_hash_unique UNIQUE (content_hash);

ALTER TABLE IF EXISTS public.checkouts
    ALTER COLUMN content_hash DROP DEFAULT;

-- Drop external_reference uniqueness — it's now purely informational
ALTER TABLE IF EXISTS public.checkouts
    DROP CONSTRAINT IF EXISTS checkouts_external_reference_unique;

COMMIT;
