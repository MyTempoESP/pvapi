CREATE FUNCTION get_product_images(p_id INT, i_tag TEXT)
RETURNS JSON AS $$
    SELECT COALESCE(json_agg(i.filename), '[]'::json)
    FROM images i
    WHERE i.product_id = p_id AND i.tag = i_tag;
$$ LANGUAGE sql STABLE;

CREATE FUNCTION get_product_policy(p_id INT, c_name TEXT)
RETURNS BOOLEAN AS $$
    SELECT c.enabled
    FROM policies c
    WHERE c.product_id = p_id AND c.name = c_name;
$$ LANGUAGE sql STABLE;

CREATE FUNCTION get_shop_products(
    p_shop_id INT,
    dtag     TEXT, -- description   images tag
    stag     TEXT, -- specification images tag
    ptag     TEXT  -- delivery      images tag
) RETURNS TABLE (
    product_id                INT,
    product_name              TEXT,
    product_price             NUMERIC,
    product_stock             INT,
    product_discount          NUMERIC,
    product_featured          BOOLEAN,
    product_moderation_status TEXT,
    product_description       TEXT,
    product_attr              JSONB,
    product_category          TEXT,

    product_specifications JSON,

	product_p_cmod   BOOLEAN,
	product_p_cname  BOOLEAN,
	product_p_ccolor BOOLEAN,

    product_platform_cut JSON, -- platform_cut object (value,type)

    product_main_image  TEXT,
    product_spec_images JSON,
    product_desc_images JSON,
    product_pack_images JSON,

    product_code TEXT,
    product_uuid UUID
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id                AS product_id,
        p.name              AS product_name,
        p.price             AS product_price,
        p.stock             AS product_stock,
        p.discount          AS product_discount,
        p.featured          AS product_featured,
        p.moderation_status AS product_moderation_status,
        p.description       AS product_description,
        p.attr              AS product_attr,
        c.name              AS product_category,

        -- =================SPECIFICATIONS=================

        -- Specifications Subquery
        COALESCE((
            SELECT json_agg(json_build_object(
                'name' , s.name,
                'value', s.value
            ))
            FROM specifications s
            WHERE s.product_id = p.id
        ), '[]'::json) AS product_specifications,

        -- ================================================

        -- ==================POLICIES=====================

        get_product_policy(p.id,    'customColor') AS product_p_ccolor,
        get_product_policy(p.id, 'customModality') AS product_p_cmod,
        get_product_policy(p.id,     'customName') AS product_p_cname,

        -- ================================================

        -- ==============EARNINGS_MODIFIERS================

        -- Platform cut needs the type info so we return everything
        (
            SELECT json_build_object(
                'type' , m.type,
                'value', m.value
            )
            FROM earnings_modifiers m
            WHERE m.product_id = p.id AND m.name = 'platform_cut'
            LIMIT 1
        ) AS product_platform_cut,

        -- ================================================

        -- ====================IMAGES======================

        -- Default image
        COALESCE((
            SELECT i.filename
            FROM images i
            WHERE i.tag = 'default' AND i.product_id = p.id
            LIMIT 1
        ), '') AS product_main_image,

        get_product_images(p.id, stag) AS product_spec_images,
        get_product_images(p.id, dtag) AS product_desc_images,
        get_product_images(p.id, ptag) AS product_pack_images,

        -- ================================================

        -- =====================ID=========================

        p.code     AS product_code,
        p.uniqueid AS product_uuid

        -- ================================================
    FROM products p
    LEFT JOIN categories c ON c.id = p.category_id
    WHERE p.shop_id = p_shop_id;
END;
$$ LANGUAGE plpgsql;
