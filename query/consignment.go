package query

const (
	CreateConsignmentQuery = `
INSERT INTO consignments
    (
	shop_id,
	pickup_location_id,
	consignee_id,
	comission_type,
	comission_value,
	status,
	code,
	slug
    )
VALUES
    (
        @shop_id,
	@pickup_location_id,
	@consignee_id,
	@comission_type,
	@comission_value,
	@status,
	uuidv4(),
	'rem-generica'
    )
RETURNING id
`

	ListAllConsignmentsQuery = `
SELECT
	consignment_id,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name,

	pickup_location_id,
	pickup_location_name
FROM consignments_listing_view
`

	FindConsignmentQuery = `
SELECT
	consignment_id,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name,

	pickup_location_id,
	pickup_location_name
FROM consignments_listing_view
WHERE consignment_id = $1
`

	DeleteConsignmentQuery = `
DELETE FROM consignments
WHERE id = $1
`

	ListAllConsignmentProductsQuery = `
SELECT
	c.product_id,
	c.product_name,
	c.product_price,
	c.product_stock,
	c.product_discount,
	c.product_featured,
	c.product_description,
	c.product_main_image,
	c.product_category_id,
	c.product_shop_id,
	c.product_code,
	c.product_uuid,

	p.price AS package_price,
	p.quantity AS package_quantity
FROM packages p
LEFT JOIN product_catalog_view c ON c.product_id = p.product_id
WHERE p.consignment_id = $1
`
)
