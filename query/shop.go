package query

import "fmt"

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

	ShopUserUpdateEmail = `
WITH shop_users AS (
	SELECT id, user_id FROM shops WHERE id = $1
)
UPDATE users
SET email = $2
FROM shop_users
WHERE users.id = shop_users.user_id
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
	SelectAllShopsQuery = fmt.Sprintf(`
SELECT
	%s
%s
`, shopSelectBaseFields, shopSelectTables)

	ShopFindQuery = fmt.Sprintf(`
%s
WHERE shop.id = $1
`, SelectAllShopsQuery)

	ShopFindByUserQuery = fmt.Sprintf(`
%s
WHERE shop.user_id = $1
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
