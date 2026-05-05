package query

const (
	CreateShipmentQuery = `
INSERT INTO shipments
(
	address_id,
	payment_id,
	client_id,
	total_amount,
	-- code: uuidv4()
	status,
	shop_id
)
SELECT
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
	@status,
	@shop_id
RETURNING id
`
)
