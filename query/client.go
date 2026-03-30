package query

const (
	CreateClientQuery = `
INSERT INTO clients
    (
        user_id
    )
VALUES
    (
        @user
    )
RETURNING id
`

	ListAllClientsQuery = `
SELECT
	u.name AS user_name,
	u.email AS user_email,
FROM clients c
LEFT JOIN users u ON u.id = c.user_id
`

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
