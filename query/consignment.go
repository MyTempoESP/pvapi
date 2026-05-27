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
  AND consignment_id IN (SELECT id FROM consignments_active)
`

	ListAllActiveConsignmentsQuery = `
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
WHERE consignment_id IN (SELECT id FROM consignments_active)
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



	VerifyPackageQuantity = `
SELECT quantity FROM packages
WHERE product_id = $1 AND consignment_id = $2
`

	ProductCheckExistsInConsignmentQuery = `
SELECT 1 FROM packages
WHERE product_id = $1 AND consignment_id = $2
LIMIT 1
`

	CheckConsignmentArchivedQuery = `
SELECT 1 FROM consignments
WHERE id = $1 AND status = 'ARQUIVADO'
LIMIT 1
`

	DeleteProductFromConsignmentQuery = `
DELETE FROM packages
WHERE product_id = $1 AND consignment_id = $2
`

	CheckPackageReservedForProductQuery = `
SELECT 1 FROM package_reserved pr
JOIN packages pkg ON pkg.id = pr.package_id
WHERE pkg.product_id = $1 AND pkg.consignment_id = $2
LIMIT 1
`
)
