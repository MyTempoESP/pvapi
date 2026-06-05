-- Revert consignment multiple pickups (pvapi-wkg5)

BEGIN;

-- Restore old column
ALTER TABLE IF EXISTS public.consignments
    ADD COLUMN IF NOT EXISTS pickup_location_id integer;

-- Migrate data back (pick first pickup if multiple)
UPDATE public.consignments c
SET pickup_location_id = (
    SELECT cpl.pickup_location_id
    FROM public.consignment_pickup_locations cpl
    WHERE cpl.consignment_id = c.id
    ORDER BY cpl.pickup_location_id
    LIMIT 1
);

-- Restore FK constraint
ALTER TABLE IF EXISTS public.consignments
    ADD CONSTRAINT consignment_pickup_location_relation FOREIGN KEY (pickup_location_id)
    REFERENCES public.pickup_locations (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE SET NULL
    NOT VALID;

-- Drop join table
DROP TABLE IF EXISTS public.consignment_pickup_locations CASCADE;

-- Restore consignments_active view (SELECT *)
CREATE OR REPLACE VIEW consignments_active AS
SELECT *
FROM consignments
WHERE status NOT IN ('ARQUIVADO', 'RASCUNHO');

COMMIT;
