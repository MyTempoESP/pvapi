BEGIN;

-- Dropping tables in reverse order of creation to respect dependencies
DROP TABLE IF EXISTS public.categories CASCADE;
DROP TABLE IF EXISTS public.earnings_modifiers CASCADE;
DROP TABLE IF EXISTS public.admins CASCADE;
DROP TABLE IF EXISTS public.policies CASCADE;
DROP TABLE IF EXISTS public.specifications CASCADE;
DROP TABLE IF EXISTS public.images CASCADE;
DROP TABLE IF EXISTS public.packages CASCADE;
DROP TABLE IF EXISTS public.products CASCADE;
DROP TABLE IF EXISTS public.reports CASCADE;
DROP TABLE IF EXISTS public.consignments CASCADE;
DROP TABLE IF EXISTS public.pickup_locations CASCADE;
DROP TABLE IF EXISTS public.consignees CASCADE;
DROP TABLE IF EXISTS public.addresses CASCADE;
DROP TABLE IF EXISTS public.shops CASCADE;
DROP TABLE IF EXISTS public.users CASCADE;

COMMIT;
