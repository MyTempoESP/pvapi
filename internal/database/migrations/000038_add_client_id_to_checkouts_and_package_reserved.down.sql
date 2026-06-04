BEGIN;

-- Drop UNIQUE(external_reference) constraint
ALTER TABLE IF EXISTS public.checkouts
    DROP CONSTRAINT IF EXISTS checkouts_external_reference_unique;

-- Drop client_id from checkouts (cascades FK drop)
ALTER TABLE IF EXISTS public.checkouts
    DROP COLUMN IF EXISTS client_id;

-- Drop client_id from package_reserved (cascades FK drop)
ALTER TABLE IF EXISTS public.package_reserved
    DROP COLUMN IF EXISTS client_id;

COMMIT;
