package query

const (
	CreateShipmentQuery = `
INSERT INTO shipments
(
	address_id,
	payment_id,
	client_id,
	total_amount,
	code,
	status,
	shop_id
)
VALUES
(
	insert_new_address(
		@addrstreet,
		@addrnumber,
		@addrneigh,
		@addrcity,
		@addrstate,
		@addrcep,
		@addrcomplement
	),
	@payment_id,
	@client_id,
	@total_amount,
	uuidv4(),
	@status,
	@shop_id
)
RETURNING id
`
)
