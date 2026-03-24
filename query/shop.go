package query

const (
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
INSERT INTO shops
	(
		name, slug, telephone,
		address_id, user_id
	)
VALUES
	(
		@name,
		@slug,
		@telephone,
		@address,
		@user
	)
RETURNING id
`

	SelectAllShopsQuery = `
SELECT * FROM shops_info_view
`

	ShopFindQuery = `
SELECT * FROM shops_info_view
WHERE shop_id = $1
`

	ShopFindByUserQuery = `
SELECT
	s.shop_id,
	s.name,
	s.slug,
	s.telephone,
	s.status,

	s.user_name,
	s.user_email,

	s.addrstreet,
	s.addrnumber,
	s.addrneigh,
	s.addrcity,
	s.addrstate,
	s.addrcep,
	s.addrcomplement
FROM shops_info_full_view s
WHERE s.user_id = $1
`

	ShopFindDetailedQuery = `
SELECT * FROM shops_info_detailed_view
WHERE shop_id = $1
`
)
