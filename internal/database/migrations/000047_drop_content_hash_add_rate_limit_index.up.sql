-- Drop content_hash dedup, add rate-limiting index (pvapi-zle6)

BEGIN;

-- Remove content_hash and its unique constraint
ALTER TABLE IF EXISTS public.checkouts
    DROP CONSTRAINT IF EXISTS checkouts_content_hash_unique;

ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS content_hash;

-- Composite index for per-client rate-limit query
CREATE INDEX IF NOT EXISTS idx_checkouts_client_processed_at
    ON public.checkouts (client_id, processed_at DESC);

COMMIT;
