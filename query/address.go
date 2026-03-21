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
)
