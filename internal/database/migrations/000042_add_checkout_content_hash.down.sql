-- Revert content_hash and restore external_reference uniqueness (pvapi-0bi4)

BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    DROP CONSTRAINT IF EXISTS checkouts_content_hash_unique;

ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS content_hash;

ALTER TABLE IF EXISTS public.checkouts
    ADD CONSTRAINT checkouts_external_reference_unique UNIQUE (external_reference);

COMMIT;
