package query

const (
	AddProductToCartQuery = `
WITH cart AS (
	INSERT INTO carts (session_id)
	VALUES (@session_id)
	ON CONFLICT (session_id) DO NOTHING
	RETURNING id
), fetch_cart AS (
	SELECT id FROM cart
	UNION ALL
	SELECT id FROM carts WHERE session_id = @session_id
	LIMIT 1
), inventory_check AS (
	SELECT product_decrement_inventory(@product_id, @quantity)
)
INSERT INTO cart_product
	(
		product_id,
		personalizations,
		quantity,
		cart_id
	)
SELECT
	@product_id,
	@personalizations,
	@quantity,
	c.id
FROM fetch_cart c
CROSS JOIN inventory_check
`

	ListAllCartsQuery = `
SELECT
    *
FROM carts
`

	FindCartQuery = `
SELECT
    *
FROM carts
WHERE id = $1
`

	DeleteCartQuery = `
DELETE FROM carts
WHERE id = $1
`
)
