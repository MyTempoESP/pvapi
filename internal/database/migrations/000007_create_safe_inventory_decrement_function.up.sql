-- Safe inventory decrement function
CREATE OR REPLACE FUNCTION product_decrement_inventory(
    product_id INTEGER,
    quantity   INTEGER
) RETURNS INTEGER AS $$
DECLARE
    current_qty INTEGER;
    new_qty     INTEGER;
BEGIN
    -- Lock the row for update (other transactions wait here)
    SELECT stock INTO current_qty
    FROM products p
    WHERE p.id = product_id
    FOR UPDATE;

    -- Check if we have enough inventory
    IF current_qty IS NULL THEN
        RAISE EXCEPTION 'Product % not found', product_id;
    END IF;

    IF current_qty < quantity THEN
        RAISE EXCEPTION 'Insufficient inventory: have %, need %',
            current_qty, quantity;
    END IF;

    -- Perform the update
    new_qty := current_qty - quantity;
    UPDATE products p
    SET p.stock = new_qty
    WHERE p.id = p_product_id;

    RETURN new_qty;
END;
$$ LANGUAGE plpgsql;
