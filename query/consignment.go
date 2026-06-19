package query

const (
	CreateConsignmentQuery = `
INSERT INTO consignments
    (
	shop_id,
	consignee_id,
	comission_type,
	comission_value,
	status,
	code,
	slug,
	name
    )
VALUES
    (
        @shop_id,
	@consignee_id,
	@comission_type,
	@comission_value,
	@status,
	uuidv4(),
	'rem-generica',
	@name
    )
RETURNING id
`

	ListConsignmentsByShopQuery = `
SELECT
	consignment_id,
	consignment_name,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view
WHERE shop_id = $1
  AND consignment_id IN (SELECT id FROM consignments_active)
`

	ListAllConsignmentsByShopQuery = `
SELECT
	consignment_id,
	consignment_name,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view
WHERE shop_id = $1
`

	ListConsignmentsByConsigneeQuery = `
SELECT
	consignment_id,
	consignment_name,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	COALESCE((SELECT SUM(pr.quantity) FROM package_reserved pr WHERE pr.consignment_id = c.consignment_id), 0) AS total_products,
	COALESCE((SELECT SUM(pr.quantity) FROM package_reserved pr WHERE pr.consignment_id = c.consignment_id AND pr.status = 'confirmed'), 0) AS confirmed_products,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view c
WHERE consignee_id = $1
  AND consignment_id IN (SELECT id FROM consignments_active)
`

	ListAllConsignmentsByConsigneeQuery = `
SELECT
	consignment_id,
	consignment_name,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	COALESCE((SELECT SUM(pr.quantity) FROM package_reserved pr WHERE pr.consignment_id = c.consignment_id), 0) AS total_products,
	COALESCE((SELECT SUM(pr.quantity) FROM package_reserved pr WHERE pr.consignment_id = c.consignment_id AND pr.status = 'confirmed'), 0) AS confirmed_products,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view c
WHERE consignee_id = $1
`

	ListAllActiveConsignmentsQuery = `
SELECT
	consignment_id,
	consignment_name,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view
WHERE consignment_id IN (SELECT id FROM consignments_active)
`


	FindConsignmentQuery = `
SELECT
	consignment_id,
	consignment_name,
	consignment_comission_type,
	consignment_comission_value,
	consignment_status,
	consignment_created_at,
	consignment_slug,
	consignment_code,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view
WHERE consignment_id = $1
`

	DeleteConsignmentQuery = `
UPDATE consignments
SET status = 'ARQUIVADO'
WHERE id = $1
`

	AddConsignmentPickupQuery = `
INSERT INTO consignment_pickup_locations (consignment_id, pickup_location_id)
VALUES ($1, $2)
ON CONFLICT (consignment_id, pickup_location_id) DO NOTHING
`

	RemoveConsignmentPickupQuery = `
DELETE FROM consignment_pickup_locations
WHERE consignment_id = $1 AND pickup_location_id = $2
`

	BatchFetchConsignmentPickupsQuery = `
SELECT
	cpl.consignment_id,
	pl.id   AS pickup_location_id,
	pl.name AS pickup_location_name
FROM consignment_pickup_locations cpl
JOIN pickup_locations pl ON pl.id = cpl.pickup_location_id
WHERE cpl.consignment_id = ANY($1)
ORDER BY pl.name
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

	FindConsignmentByCodeQuery = `
SELECT
	id,
	status
FROM consignments
WHERE code = $1
`

	PickupLocationsByConsignmentCodeQuery = `
SELECT
	pl.id               AS pickup_id,
	pl.name AS pickup_name,
	a.addrstreet,
	a.addrnumber::text AS addrnumber,
	a.addrneigh,
	a.addrcity,
	a.addrstate,
	a.addrcep,
	a.addrcomplement
FROM consignment_pickup_locations cpl
JOIN pickup_locations pl ON pl.id = cpl.pickup_location_id
JOIN aliased_addresses_view a ON a.address_id = pl.address_id
WHERE cpl.consignment_id = (
	SELECT id FROM consignments WHERE code = $1
)
  AND pl.valid_until > NOW()
`

)
