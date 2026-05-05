package query

const (
	CreateShipmentQuery = `
INSERT INTO shipments
(
	payment_id,
	address_id,
	cart_id,
	status,
	shop_id
)
SELECT DISTINCT
	py.id,
	py.address_id,
	py.cart_id,
	'pending',
	p.shop_id
FROM payments py
JOIN cart_product cp ON cp.cart_id = py.cart_id
JOIN products p ON p.id = cp.product_id
WHERE py.payment_id = $1
RETURNING id
`
)
