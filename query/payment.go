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
	payment_method,
	currency,
	external_reference,
	method,
	metadata
)
VALUES
(
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
	@payment_method,
	@currency,
	@external_reference,
	@method,
	@metadata
)
`
)