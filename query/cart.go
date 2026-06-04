package query

const (
	FindActiveCartBySessionQuery = `
SELECT id
FROM carts
WHERE session_id = $1
  AND session_id NOT LIKE 'FINISHED-%'
`

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
RETURNING id
`

	InsertCartProductQuery = `
INSERT INTO cart_product
	(
		product_id,
		personalizations,
		quantity,
		cart_id
	)
VALUES
	(
		$1,
		$2,
		$3,
		$4
	)
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
USING carts c
WHERE cart_product.id = $1 AND c.id = cart_product.cart_id AND c.session_id = $2
`

	GetCartItemProductID = `
SELECT product_id FROM cart_product
JOIN carts c ON c.id = cart_product.cart_id
WHERE cart_product.id = $1 AND c.session_id = $2
`


)

