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
	status
    )
VALUES
    (
        @shop_id,
	@pickup_location_id,
	@consignee_id,
	@comission_type,
	@comission_value,
	@status
    )
RETURNING id
`

	ListAllConsignmentsQuery = `
SELECT
	consignment_id,
	consignment_comission_type,
	consignment_comission_value,
	consignment_pickup_location,
	consignment_status,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view
`

	FindConsignmentQuery = `
SELECT
	consignment_id,
	consignment_comission_type,
	consignment_comission_value,
	consignment_pickup_location,
	consignment_status,

	consignee_id,
	consignee_name,

	shop_id,
	shop_name
FROM consignments_listing_view
WHERE consignment_id = $1
`

	DeleteConsignmentQuery = `
DELETE FROM consignments
WHERE id = $1
`
)
