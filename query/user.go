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

	// level
	UserMaxLevelQuery = `
SELECT
	(admins.id IS NOT NULL) AS admin,
	(shops.id IS NOT NULL) AS shop,
	(consignees.id IS NOT NULL) AS consignee,
	users.id
FROM users
LEFT JOIN admins ON users.id = admins.user_id
LEFT JOIN shops ON users.id = shops.user_id
LEFT JOIN consignees ON users.id = consignees.user_id
WHERE users.id = $1
`
)
