package query

const (
	PaymentCreateQuery = `
INSERT INTO payments
(
	cart_id,
	transaction_amount,
	quantity,
	amount,
	status,
	description,
	title,
	payment_id,
	payer_name,
	payer_email,
	payer_surname,
	address_id,
	payment_method,
	currency,
	external_reference,
	method,
	metadata
)
SELECT
	@cart_id,
	@transaction_amount,
	@quantity,
	@amount,
	@status,
	@description,
	@title,
	@payment_id,
	@payer_name,
	@payer_email,
	@payer_surname,
	insert_new_address(
		@addrstreet,
		@addrnumber,
		@addrneigh,
		@addrcity,
		@addrstate,
		@addrcep,
		@addrcomplement
	),
	@payment_method,
	@currency,
	@external_reference,
	@method,
	@metadata
RETURNING id
`
)

