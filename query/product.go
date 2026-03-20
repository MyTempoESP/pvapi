package query

const (
	// creating
	ProductCreate = `
WITH product_category AS (
	INSERT INTO categories
	(
		name,
		description,
		shop_id
	)
	VALUES
	(
		@category,
		'',
		@shop_id
	)
	ON CONFLICT DO NOTHING
	RETURNING id
)
INSERT INTO products
	(
		name,
		price,
		stock,
		discount,
		featured,
		shop_id,
		moderation_status,
		category_id,
		description,
		attr,
		code,
		uniqueid
	)
VALUES
	(
		@name,
		@price,
		@stock,
		@discount,
		@featured,
		@shop_id,
		'pending',
		(SELECT id FROM product_category),
		@description,
		@attr,
		@code,
		uuidv4()
	)
RETURNING id
`

	// listing
	ProductGetByShop = `
SELECT * FROM get_shop_products(@shop_id, @desc_tag, @spec_tag, @pack_tag)
`
)
