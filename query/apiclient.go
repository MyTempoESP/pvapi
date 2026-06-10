package query

const (
	CreateAPIClientQuery = `
INSERT INTO api_clients
	(name, email, token_hash)
VALUES
	($1, $2, $3)
RETURNING id
`

	ListAPIClientsQuery = `
SELECT
	id,
	name,
	email,
	token_hash,
	created_at,
	revoked_at
FROM api_clients
ORDER BY created_at DESC
`

	FindAPIClientByIDQuery = `
SELECT
	id,
	name,
	email,
	token_hash,
	created_at,
	revoked_at
FROM api_clients
WHERE id = $1
`

	FindAPIClientByTokenHashQuery = `
SELECT
	id,
	name,
	email,
	token_hash,
	created_at,
	revoked_at
FROM api_clients
WHERE token_hash = $1
`

	RevokeAPIClientQuery = `
UPDATE api_clients
SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL
RETURNING id
`
)
