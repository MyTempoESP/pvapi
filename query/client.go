package query

const (
	CreateClientQuery = `
INSERT INTO clients
    (
        user_id
    )
VALUES
    (
        $1
    )
RETURNING id
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
