package query

const (
	CreatePickupLocationQuery = `
INSERT INTO pickup_locations
	(
		name,
		pickup_at,
		address_id,
		shop_id
	)
	VALUES
	(
		@name,
		@pickup_at,
		insert_new_address(
			@addrstreet,
			@addrnumber,
			@addrneigh,
			@addrcity,
			@addrstate,
			@addrcep,
			@addrcomplement
		),
		@shop_id
	)
RETURNING id
`

	ListPickupLocationsQuery = `
SELECT
	p.id,
	p.name,
	p.pickup_at,
	p.address_id
FROM pickup_locations p
WHERE p.shop_id = $1
`
)
