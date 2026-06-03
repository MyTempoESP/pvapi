BEGIN;

-- Revert product_catalog_view to previous definition (with package_reserved subquery)
DROP VIEW IF EXISTS public.product_catalog_view;
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

-- Remove denormalized columns
ALTER TABLE IF EXISTS public.package_reserved
    DROP COLUMN IF EXISTS shop_id,
    DROP COLUMN IF EXISTS consignment_id,
    DROP COLUMN IF EXISTS shop_value,
    DROP COLUMN IF EXISTS status;

-- Restore FK constraint
ALTER TABLE IF EXISTS public.package_reserved
    DROP CONSTRAINT IF EXISTS package_reserved_package_relation;

ALTER TABLE IF EXISTS public.package_reserved
    ALTER COLUMN package_id SET NOT NULL;

ALTER TABLE IF EXISTS public.package_reserved
    ADD CONSTRAINT package_reserved_package_relation FOREIGN KEY (package_id)
    REFERENCES public.packages (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

-- Drop the function
DROP FUNCTION IF EXISTS package_decrement_quantity(INTEGER, INTEGER);

COMMIT;
