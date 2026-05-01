package query

const (
	CreateClientQuery = `
WITH client AS (
	INSERT INTO clients (user_id)
	VALUES ($1)
	ON CONFLICT (user_id) DO NOTHING
	RETURNING id
)
SELECT id FROM client
UNION ALL
SELECT id FROM clients WHERE user_id = $1
LIMIT 1
`

	ClientAssociateToCartQuery = `
UPDATE carts SET client_id = $1 WHERE id = $2
`

	ListShopClientsQuery = `
SELECT
	u.name AS user_name,
	u.email AS user_email
FROM clients c
LEFT JOIN users u ON u.id = c.user_id
WHERE c.shop_id = $1
`

	// TODO: only way to find clients is through payments
	// this is wrong.
	FindClientQuery = `
SELECT
	u.name AS user_name,
	u.email AS user_email,
FROM clients c
LEFT JOIN users u ON u.id = c.user_id
WHERE id = $1
`

	DeleteClientQuery = `
DELETE FROM clients
WHERE id = $1
`
)
