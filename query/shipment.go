package query

const (
	CreateShipmentQuery = `
INSERT INTO shipments
(
	payment_id,
	address_id,
	client_id,
	cart_id,
	status,
	shop_id
)
SELECT DISTINCT
	p.id,
	p.address_id,
	p.client_id,
	p.cart_id,
	'pending',
	c.shop_id
FROM payments p
JOIN cart_products c ON c.cart_id = p.cart_id
WHERE p.id = $1
RETURNING id
`
)
