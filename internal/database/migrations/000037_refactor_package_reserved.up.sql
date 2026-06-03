BEGIN;

-- 1. Create package_decrement_quantity function
-- Atomically checks and decrements packages.quantity.
CREATE OR REPLACE FUNCTION package_decrement_quantity(
    p_id INTEGER,
    p_qty INTEGER
) RETURNS INTEGER AS $$
DECLARE
    current_qty INTEGER;
    new_qty     INTEGER;
BEGIN
    SELECT quantity INTO current_qty
    FROM packages
    WHERE id = p_id
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Package % not found', p_id;
    END IF;

    IF current_qty < p_qty THEN
        RAISE EXCEPTION 'Insufficient stock: have %, need %', current_qty, p_qty;
    END IF;

    new_qty := current_qty - p_qty;

    UPDATE packages
    SET quantity = new_qty
    WHERE id = p_id;

    RETURN new_qty;
END;
$$ LANGUAGE plpgsql;

-- 2. Alter package_reserved.package_id FK: ON DELETE SET NULL, column nullable
ALTER TABLE IF EXISTS public.package_reserved
    DROP CONSTRAINT IF EXISTS package_reserved_package_relation;

ALTER TABLE IF EXISTS public.package_reserved
    ALTER COLUMN package_id DROP NOT NULL;

ALTER TABLE IF EXISTS public.package_reserved
    ADD CONSTRAINT package_reserved_package_relation FOREIGN KEY (package_id)
    REFERENCES public.packages (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE SET NULL
    NOT VALID;

-- 3. Add denormalized columns to package_reserved
ALTER TABLE IF EXISTS public.package_reserved
    ADD COLUMN IF NOT EXISTS shop_id INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS consignment_id INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS shop_value NUMERIC(12, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';

-- Drop default after adding columns (not null columns need defaults during add)
ALTER TABLE IF EXISTS public.package_reserved
    ALTER COLUMN shop_id DROP DEFAULT,
    ALTER COLUMN consignment_id DROP DEFAULT,
    ALTER COLUMN shop_value DROP DEFAULT;

-- 4. Update product_catalog_view: replace subquery with SUM(pkg.quantity)
DROP VIEW IF EXISTS public.product_catalog_view;
CREATE OR REPLACE VIEW public.product_catalog_view AS
SELECT
    p.id                AS product_id,
    p.name              AS product_name,
    p.price             AS product_price,
    COALESCE(
        (SELECT SUM(pkg.quantity)
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
