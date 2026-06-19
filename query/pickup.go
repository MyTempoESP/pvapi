package query

const (
	CreatePickupLocationQuery = `
INSERT INTO pickup_locations
	(
		name,
		description,
		starts_at,
		valid_until,
		address_id,
		shop_id
	)
	VALUES
	(
		@name,
		@description,
		@starts_at,
		@valid_until,
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
	p.description,
	p.starts_at,
	p.valid_until,

	a.addrstreet,
	a.addrnumber,
	a.addrcity,
	a.addrstate,
	a.addrcep,
	a.addrneigh,
	a.addrcomplement
FROM pickup_locations p
LEFT JOIN aliased_addresses_view a USING (address_id)
WHERE p.shop_id = $1
`

	FindPickupLocationQuery = `
SELECT
	p.id,
	p.name,
	p.description,
	p.starts_at,
	p.valid_until,

	a.addrstreet,
	a.addrnumber,
	a.addrcity,
	a.addrstate,
	a.addrcep,
	a.addrneigh,
	a.addrcomplement
FROM pickup_locations p
LEFT JOIN aliased_addresses_view a USING (address_id)
WHERE p.id = $1
`

	UpdatePickupLocationQuery = `
UPDATE pickup_locations
SET name   = COALESCE($2::text, name),
    starts_at = COALESCE($3::timestamptz, starts_at),
    description = COALESCE($4::text, description),
    valid_until  = COALESCE($5::timestamptz, valid_until)
WHERE id = $1
`

	DeletePickupLocationQuery = `
DELETE FROM pickup_locations WHERE id = $1
`
)
