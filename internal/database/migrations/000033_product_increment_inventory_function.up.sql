CREATE OR REPLACE FUNCTION product_increment_inventory(
    p_id INTEGER,
    p_qty INTEGER
) RETURNS INTEGER AS $$
DECLARE
    current_qty INTEGER;
    new_qty     INTEGER;
BEGIN
    -- 1. Lock and Select
    SELECT stock INTO current_qty
    FROM products
    WHERE id = p_id
    FOR UPDATE;

    -- 2. Validate
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Product % not found', p_id;
    END IF;

    -- 3. Calculate and Update
    new_qty := current_qty + p_qty;

    UPDATE products
    SET stock = new_qty
    WHERE id = p_id;

    RETURN new_qty;
END;
$$ LANGUAGE plpgsql;
