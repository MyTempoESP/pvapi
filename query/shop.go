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
WITH new_shop AS (
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
			insert_new_address(
				@addrstreet,
				@addrnumber,
				@addrneigh,
				@addrcity,
				@addrstate,
				@addrcep,
				@addrcomplement
			),
			@user
		)
	RETURNING id
)
INSERT INTO categories
	(
		name,
		description,
		shop_id
	)
SELECT
	'n/a',
	'Sem categoria',
	s.id
FROM new_shop s
RETURNING shop_id
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
SELECT
    shop_id,
    name,
    slug,
    telephone,
    status,

    user_name,
    user_email,

    addrstreet,
    addrnumber,
    addrneigh,
    addrcity,
    addrstate,
    addrcep,
    addrcomplement,

    products_count,
    pickups
FROM shops_info_detailed_view
WHERE shop_id = $1
`

	// This query uses 'v' to mask one of the outputs (v.user_id)
	// which we do not want to return.
	ShopFindDetailedByUserQuery = `
SELECT
    v.shop_id,
    v.name,
    v.slug,
    v.telephone,
    v.status,

    v.user_name,
    v.user_email,
    v.addrstreet,
    v.addrnumber,
    v.addrneigh,
    v.addrcity,
    v.addrstate,
    v.addrcep,
    v.addrcomplement,

    v.products_count,
    v.pickups
FROM shops_info_detailed_view v
WHERE v.user_id = $1
`
)
