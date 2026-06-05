-- Add unique_id to package_reserved for per-package QR codes (pvapi-lgua)

BEGIN;

ALTER TABLE IF EXISTS public.package_reserved
    ADD COLUMN IF NOT EXISTS unique_id UUID NOT NULL DEFAULT gen_random_uuid();

-- Keep the default so that INSERT statements that omit unique_id auto-generate it.

COMMIT;
