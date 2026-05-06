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
	'PENDENTE SEPARAÇÃO',
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
  AND consignment_status != 'ARQUIVADO'
ORDER BY consignment_created_at DESC
`

	ListShipmentProductsQuery = `
SELECT
    c.product_id,
    c.product_name,
    c.product_price,
    c.product_stock,
    c.product_discount,
    c.product_featured,
    c.product_moderation_status,
    c.product_description,
    c.product_category_id,
    c.product_category,
    c.product_uuid,
    c.product_code,
    c.product_attr,
    c.product_platform_cut,
    c.product_platform_cut_type,
    c.product_shop_id,
    c.product_shop_name,

    cp.price        AS package_price,
    cp.quantity     AS package_quantity
FROM shipments sh
JOIN cart_product cp ON cp.cart_id = sh.cart_id
JOIN product_catalog_view c ON c.product_id = cp.product_id
WHERE sh.id = $1 AND sh.shop_id = c.product_shop_id
`
)

