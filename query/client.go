package query

const (
	CreateClientQuery = `
INSERT INTO clients (user_id, telephone)
VALUES ($1, $2)
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
SELECT DISTINCT
	u.id AS user_id,
	u.name AS user_name,
	u.email AS user_email,
	c.telephone AS user_phone
FROM clients c
LEFT JOIN users u ON u.id = c.user_id
LEFT JOIN carts cc ON cc.client_id = c.id
LEFT JOIN cart_product cp ON cp.cart_id = cc.id
LEFT JOIN products p ON p.id = cp.product_id
WHERE p.shop_id = $1
`

	FindClientByUserQuery = `
SELECT id FROM clients WHERE user_id = $1
`

	DeleteClientQuery = `
DELETE FROM clients
WHERE id = $1
`
)
