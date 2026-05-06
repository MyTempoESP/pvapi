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
  AND consignment_status != 'ARQUIVADO'
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
UPDATE consignments
SET status = 'ARQUIVADO'
WHERE id = $1
`

	FindFromViewQuery = `
SELECT *
FROM consignment_shipment_view
WHERE consignment_id = $1 AND record_type = $2
`

	VerifyPackageQuantity = `
SELECT quantity FROM packages
WHERE product_id = $1 AND consignment_id = $2
`
)
