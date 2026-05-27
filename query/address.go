package query

const (
	CreateAddressQuery = `
INSERT into addresses
(
	street, number, neighborhood, city,
	state, cep, complement
)
VALUES
(
	@street, @number,
	@neigh, @city,
	@state, @cep,
	@complement
)
RETURNING id
`

	UpdateAddressQuery = `
UPDATE addresses
SET street      = COALESCE($2::text, street),
    number      = COALESCE($3::int, number),
    neighborhood = COALESCE($4::text, neighborhood),
    city        = COALESCE($5::text, city),
    state       = COALESCE($6::text, state),
    cep         = COALESCE($7::text, cep),
    complement  = COALESCE($8::text, complement)
WHERE id = (
	SELECT address_id FROM pickup_locations WHERE id = $1
)
`
)
