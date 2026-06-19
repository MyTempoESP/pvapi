-- Revert consignment statuses (pvapi-piqn)

BEGIN;

DROP VIEW IF EXISTS consignments_active CASCADE;

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

COMMIT;
