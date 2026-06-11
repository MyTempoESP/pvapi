package query

const (
	CreateClientQuery = `
INSERT INTO clients (user_id, telephone, cpf)
VALUES ($1, NULLIF($2, ''), NULLIF($3, ''))
ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
RETURNING id
`

	UpdateClientFieldsQuery = `
UPDATE clients
SET
	telephone = CASE WHEN $2::TEXT != '' THEN $2::TEXT ELSE telephone END,
	cpf       = CASE WHEN $3::TEXT != '' THEN $3::TEXT ELSE cpf END
WHERE id = $1
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

	LockClientRowQuery = `
SELECT id FROM clients WHERE id = $1 FOR UPDATE
`

	DeleteClientQuery = `
DELETE FROM clients
WHERE id = $1
`
)
