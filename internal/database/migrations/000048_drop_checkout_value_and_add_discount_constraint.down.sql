-- Restore value and value_discrepancy columns, drop discount constraint.

BEGIN;

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN value numeric(12, 2) NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS public.checkouts
    ADD COLUMN value_discrepancy numeric(12, 2) NOT NULL DEFAULT 0;

ALTER TABLE IF EXISTS public.packages
    DROP CONSTRAINT IF EXISTS discount_range;

COMMIT;
