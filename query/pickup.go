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

	a.addrstreet,
	a.addrnumber,
	a.addrneigh,
	a.addrcity,
	a.addrstate,
	a.addrcep,
	a.addrcomplement
FROM pickup_locations p
LEFT JOIN addresses a ON a.id = p.address_id
WHERE p.shop_id = $1
`
)
