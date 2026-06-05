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

-- Restore consignments_listing_view with pickup columns
DROP VIEW IF EXISTS consignments_listing_view;
CREATE OR REPLACE VIEW consignments_listing_view AS
SELECT
    c.id              AS consignment_id,
    c.comission_type  AS consignment_comission_type,
    c.comission_value AS consignment_comission_value,
    c.status          AS consignment_status,
    c.created_at      AS consignment_created_at,
    c.slug            AS consignment_slug,
    c.code            AS consignment_code,

    m.consignee_id AS consignee_id,
    m.user_name    AS consignee_name,

    s.shop_id AS shop_id,
    s.name    AS shop_name,

    p.id   AS pickup_location_id,
    p.name AS pickup_location_name
FROM consignments c
LEFT JOIN shops_info_view      s USING (shop_id)
LEFT JOIN consignees_info_view m USING (consignee_id)
LEFT JOIN pickup_locations     p ON p.id = c.pickup_location_id
;

COMMIT;
