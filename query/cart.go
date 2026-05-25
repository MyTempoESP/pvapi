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

	ListCartConsignmentOptions = `
SELECT
	cp.product_id,
	p.name AS product_name,

	c.id AS consignment_id,
	c.status AS consignment_status,
	c.code AS consignment_code,

	pkg.price AS consignment_price,
	pkg.quantity AS consignment_stock,
	pkg.discount AS consignment_discount,

	pl.id AS pickup_id,
	pl.name AS pickup_name,
	pl.pickup_at AS pickup_date,

	COALESCE(a.addrstreet, '') AS addrstreet,
	COALESCE(a.addrnumber::text, '') AS addrnumber,
	COALESCE(a.addrcity, '')    AS addrcity,
	COALESCE(a.addrstate, '')   AS addrstate,
	COALESCE(a.addrcep, '')     AS addrcep,
	COALESCE(a.addrneigh, '')   AS addrneigh,
	COALESCE(a.addrcomplement, '') AS addrcomplement
FROM cart_product cp
JOIN carts cart ON cart.id = cp.cart_id
JOIN products p ON p.id = cp.product_id
LEFT JOIN packages pkg ON pkg.product_id = cp.product_id
LEFT JOIN consignments c ON c.id = pkg.consignment_id
	AND c.status NOT IN ('ARQUIVADO', 'RASCUNHO')
LEFT JOIN pickup_locations pl ON pl.id = c.pickup_location_id
LEFT JOIN aliased_addresses_view a ON a.address_id = pl.address_id
WHERE cart.session_id = $1
ORDER BY cp.product_id, c.id
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

