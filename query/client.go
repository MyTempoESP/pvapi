package query

const (
	CreateClientQuery = `
INSERT INTO clients (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING id
`

	ClientAssociateToCartQuery = `
UPDATE carts SET
	client_id = $1,
	session_id = 'FINISHED_' || uuidv4()::text
WHERE id = $2
`

	ListShopClientsQuery = `
SELECT
	u.name AS user_name,
	u.email AS user_email
FROM clients c
LEFT JOIN users u ON u.id = c.user_id
LEFT JOIN carts cc ON cc.client_id = c.id
LEFT JOIN shipments s ON s.cart_id = cc.id
WHERE s.shop_id = $1
`

	// TODO: only way to find clients is through payments
	// this is wrong.
	FindClientQuery = `
SELECT
	u.name AS user_name,
	u.email AS user_email
FROM clients c
LEFT JOIN users u ON u.id = c.user_id
WHERE id = $1
`

	DeleteClientQuery = `
DELETE FROM clients
WHERE id = $1
`
)
