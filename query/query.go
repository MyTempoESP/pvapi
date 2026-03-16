package query

import "fmt"

// Building blocks
const (
	// Shop fields
	shopSelectBaseFields = `
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
`
	shopSelectTables = `
FROM shops AS shop
LEFT JOIN addresses AS address ON shop.address_id = address.id
LEFT JOIN users AS u ON shop.user_id = u.id
`
)

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

// Expansion queries
var (
	SelectAllShopsQuery = fmt.Sprintf(`
SELECT
	%s
%s
`, shopSelectBaseFields, shopSelectTables)

	ShopFindQuery = fmt.Sprintf(`
%s
WHERE shop.id = $1
`, SelectAllShopsQuery)

	ShopFindDetailedQuery = fmt.Sprintf(`
WITH product_count AS (
	SELECT shop_id, COUNT(id) AS count
	FROM products
	GROUP BY shop_id
),
pickup_count AS (
	SELECT shop_id, COUNT(id) AS count
	FROM pickup_locations
	GROUP BY shop_id
)
SELECT
	%s,
	COALESCE(product.count, 0) AS products_count,
	COALESCE(pickup.count, 0) AS pickups
%s
LEFT JOIN product_count AS product ON product.shop_id = shop.id
LEFT JOIN pickup_count AS pickup ON pickup.shop_id = shop.id
WHERE shop.id = $1
`, shopSelectBaseFields, shopSelectTables)
)
