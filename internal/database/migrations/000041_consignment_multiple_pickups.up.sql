-- Replace single consignments.pickup_location_id FK with a join table
-- for multiple pickup locations per consignment (pvapi-wkg5)

BEGIN;

-- Drop all views that depend on consignments.pickup_location_id
DROP VIEW IF EXISTS consignments_active CASCADE;
DROP VIEW IF EXISTS consignments_listing_view;

-- Recreate consignments_active without pickup_location_id
CREATE VIEW consignments_active AS
SELECT id, shop_id, consignee_id, comission_type, comission_value,
       created_at, status, code, slug
FROM consignments
WHERE status NOT IN ('ARQUIVADO', 'RASCUNHO');

-- Recreate product_catalog_view (dependency dropped by CASCADE above)
CREATE OR REPLACE VIEW public.product_catalog_view AS
SELECT
    p.id                AS product_id,
    p.name              AS product_name,
    p.price             AS product_price,
    COALESCE(
        (SELECT SUM(pkg.quantity - COALESCE(
            (SELECT SUM(pr.quantity)
             FROM package_reserved pr
             WHERE pr.package_id = pkg.id), 0
        ))
         FROM packages pkg
         JOIN consignments_active con ON con.id = pkg.consignment_id
         WHERE pkg.product_id = p.id), 0
    )::integer          AS product_stock,
    p.discount          AS product_discount,
    p.featured          AS product_featured,
    p.moderation_status AS product_moderation_status,
    p.description       AS product_description,
    p.shop_id           AS product_shop_id,
    s.name              AS product_shop_name,
    p.attr              AS product_attr,
    c.name              AS product_category,
    c.id                AS product_category_id,
    platform_cut        AS product_platform_cut,
    platform_cut_type   AS product_platform_cut_type,
    p.code              AS product_code,
    p.uniqueid          AS product_uuid
FROM products p
LEFT JOIN shops      s ON s.id = p.shop_id
LEFT JOIN categories c ON c.id = p.category_id;

-- Recreate product_catalog_shop_view (also a dependency)
CREATE OR REPLACE VIEW public.product_catalog_shop_view AS
SELECT
    p.id                AS product_id,
    p.name              AS product_name,
    p.price             AS product_price,
    p.stock             AS product_stock,
    p.discount          AS product_discount,
    p.featured          AS product_featured,
    p.moderation_status AS product_moderation_status,
    p.description       AS product_description,
    p.shop_id           AS product_shop_id,
    s.name              AS product_shop_name,
    p.attr              AS product_attr,
    c.name              AS product_category,
    c.id                AS product_category_id,
    platform_cut        AS product_platform_cut,
    platform_cut_type   AS product_platform_cut_type,
    p.code              AS product_code,
    p.uniqueid          AS product_uuid
FROM products p
LEFT JOIN shops      s ON s.id = p.shop_id
LEFT JOIN categories c ON c.id = p.category_id;

-- Create join table
CREATE TABLE IF NOT EXISTS public.consignment_pickup_locations (
    consignment_id    integer NOT NULL,
    pickup_location_id integer NOT NULL,
    PRIMARY KEY (consignment_id, pickup_location_id),
    CONSTRAINT cpl_consignment_fk FOREIGN KEY (consignment_id)
        REFERENCES public.consignments (id)
        ON DELETE CASCADE,
    CONSTRAINT cpl_pickup_location_fk FOREIGN KEY (pickup_location_id)
        REFERENCES public.pickup_locations (id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cpl_pickup_location
    ON public.consignment_pickup_locations(pickup_location_id);

-- Migrate existing rows
INSERT INTO public.consignment_pickup_locations (consignment_id, pickup_location_id)
SELECT id, pickup_location_id
FROM public.consignments
WHERE pickup_location_id IS NOT NULL;

-- Recreate consignments_listing_view without pickup_location_id
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
    s.name    AS shop_name
FROM consignments c
LEFT JOIN shops_info_view      s USING (shop_id)
LEFT JOIN consignees_info_view m USING (consignee_id)
;

-- Remove old FK constraint
ALTER TABLE IF EXISTS public.consignments
    DROP CONSTRAINT IF EXISTS consignment_pickup_location_relation;

-- Drop old column
ALTER TABLE IF EXISTS public.consignments
    DROP COLUMN IF EXISTS pickup_location_id;

COMMIT;
