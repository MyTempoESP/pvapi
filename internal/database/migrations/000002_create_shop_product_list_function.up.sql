CREATE FUNCTION public.get_product_images(p_id INT, i_tag TEXT)
RETURNS JSON AS $$
    SELECT COALESCE(json_agg(i.filename), '[]'::json)
    FROM images i
    WHERE i.product_id = p_id AND i.tag = i_tag;
$$ LANGUAGE sql STABLE;

CREATE VIEW public.product_catalog_view AS
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

    -- ====================IMAGES======================
    COALESCE((
        SELECT i.filename
        FROM images i
        WHERE i.tag = 'default' AND i.product_id = p.id
        LIMIT 1
    ), '') AS product_main_image,

    get_product_images(p.id, 'specification') AS product_spec_images,
    get_product_images(p.id,   'description') AS product_desc_images,
    get_product_images(p.id,      'delivery') AS product_pack_images,
    -- ================================================

    -- =====================ID=========================
    p.code     AS product_code,
    p.uniqueid AS product_uuid
    -- ================================================
FROM products p
LEFT JOIN shops      s ON s.id = p.shop_id
LEFT JOIN categories c ON c.id = p.category_id;
