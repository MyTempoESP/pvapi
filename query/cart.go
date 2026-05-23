package query

const (
	CreateCartQuery = `
INSERT INTO carts
	(
		client_id,
		session_id
	)
VALUES
	(
		NULL,
		$1
	)
`

	AddProductToCartQuery = `
WITH fetch_cart AS (
	SELECT id FROM carts WHERE session_id = @session_id
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
`

	VerifyCartItemQuantity = `
SELECT quantity
FROM cart_product
WHERE product_id = $1 AND id = $2
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

	CartFindBySession = `
SELECT
	*
FROM carts
WHERE session_id = $1
`

	DeleteCartQuery = `
DELETE FROM carts
WHERE id = $1
`

	ListCartProducts = `
SELECT
	cp.id,
	cp.product_id,
	cp.cart_id,
	cp.personalizations,
	cp.quantity,

	p.id AS cart_product_id,
	p.name AS cart_product_name,
	p.price AS cart_product_price,
	p.attr->'main_image' AS
		cart_product_image,
	p.discount AS
		cart_product_discount,
	p.shop_id AS cart_product_shop,
	s.name AS cart_product_shop_name
FROM cart_product cp
LEFT JOIN products p ON
	p.id = cp.product_id
LEFT JOIN shops s ON
	s.id = p.shop_id
WHERE cart_id = $1
`

	RemoveCartProductQuery = `
DELETE FROM cart_product
JOIN carts c ON c.id = cart_id
WHERE id = $1 AND c.session_id = $2
`

	GetCartItemProductID = `
SELECT product_id FROM cart_product
JOIN carts c ON c.id = cart_id
WHERE id = $1 AND c.session_id = $2
`

	// freeze prices & hold stock
	CartHoldStock = `
WITH update_cart_products AS (
	UPDATE cart_product i
	SET price = p.price * (1 - p.discount / 100) * i.quantity
	FROM products p
	WHERE i.cart_id = $1
	  AND i.product_id = p.id
)
SELECT product_decrement_inventory(product_id, quantity)
FROM cart_product
WHERE cart_id = $1
`
)

