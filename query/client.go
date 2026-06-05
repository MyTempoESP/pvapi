package query

const (
	CreateClientQuery = `
INSERT INTO clients (user_id, telephone)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING id
`

	ListShopClientsQuery = `
SELECT
	u.id              AS user_id,
	u.name            AS user_name,
	u.email           AS user_email,
	c.telephone       AS user_phone,
	COALESCE(SUM(cs.shop_value), 0)            AS total_purchases,
	MAX(cs.created_at)                          AS last_purchase
FROM clients c
JOIN users u ON u.id = c.user_id
JOIN checkout_shop cs ON cs.client_id = c.id
WHERE cs.shop_id = $1
GROUP BY u.id, u.name, u.email, c.telephone
`

	FindClientByUserQuery = `
SELECT id FROM clients WHERE user_id = $1
`

	DeleteClientQuery = `
DELETE FROM clients
WHERE id = $1
`
)
