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

	ListConsignmentsByShopQuery = `
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
WHERE shop_id = $1
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
)
