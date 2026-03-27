package query

const (
	// seeding
	AdminCreateQuery = `
INSERT INTO admins (user_id) VALUES ($1)
`

	// Auth
	UserPasswordQuery = `
SELECT id, password_hash FROM users WHERE email = $1
`
	UserRegisterQuery = `
INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3)
RETURNING id
`

	// user
	UserDataQuery = `
SELECT name, email FROM users WHERE id = $1
`
	UserFindByEmailQuery = `
SELECT id, email FROM users WHERE email = $1
`

	UserFindOrCreateQuery = `
WITH inserted_user AS (
	INSERT INTO users
	(
		name,
		email,
		password_hash
	)
	VALUES
	(
		@name,
		@email,
		@password
	)
	ON CONFLICT (email) DO NOTHING
	RETURNING id
)
SELECT id FROM inserted_user
UNION ALL
SELECT id FROM users u WHERE u.email = @email
LIMIT 1
`

	// level
	UserPermissionsQuery = `
SELECT
	a.id AS admin,
	s.id AS shop,
	c.id AS consignee,
	l.id AS client
FROM users           u
LEFT JOIN admins     a ON u.id = a.user_id
LEFT JOIN shops      s ON u.id = s.user_id
LEFT JOIN consignees c ON u.id = c.user_id
LEFT JOIN clients    l ON u.id = l.user_id
WHERE u.id = $1
`
)
