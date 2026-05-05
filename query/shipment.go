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

	// ListConsignmentsWithShipmentsQuery uses the unified view.
	// $1 is shop_id
	ListConsignmentsWithShipmentsQuery = `
		SELECT *
		FROM consignment_shipment_view
		WHERE shop_id = $1
		ORDER BY consignment_created_at DESC
	`
)
