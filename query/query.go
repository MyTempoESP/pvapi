package query

import "fmt"

const (
	// Auth
	UserPasswordQuery = `
SELECT id, password_hash FROM users WHERE email = $1
`
	UserRegisterQuery = `
INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3)
RETURNING id
`

	// general
	UserDataQuery = `
SELECT name, email FROM users WHERE id = $1
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

	// seeding
	AdminCreateQuery = `
INSERT INTO admins (user_id) VALUES ($1)
`

	// Listing
	SelectAllShopsQuery = `
SELECT
	shop.id AS shop_id,
	shop.name AS name,
	shop.slug AS slug,
	shop.telephone AS telephone,
	'active' AS status,

	u.name AS user_name,
	u.email AS user_email,

	address.street AS addrstreet,
	address.number AS addrnumber,
	address.neighborhood AS addrneigh,
	address.city AS addrcity,
	address.state AS addrstate,
	address.cep AS addrcep,
	address.complement AS addrcomplement
FROM shops AS shop
LEFT JOIN addresses AS address ON shop.address_id = address.id
LEFT JOIN users AS u ON shop.user_id = u.id
`

	ShopCreateQuery = `
WITH new_address AS (
	INSERT into addresses
	(
		street, number, neighborhood, city,
		state, cep, complement
	)
	VALUES
	(
		@addrstreet, @addrnumber,
		@addrneigh, @addrcity,
		@addrstate, @addrcep,
		@addrcomplement
	)

	returning id
)
INSERT INTO shops
	(
		name, slug, telephone,
		address_id, user_id
	)
SELECT
	@name,
	@slug,
	@telephone,
	new_address.id,
	@user
FROM new_address
RETURNING id
`
)

var (
	ShopFindQuery = fmt.Sprintf(`
%s
WHERE shop.id = $1
`, SelectAllShopsQuery)
)
