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
	ON CONFLICT (shop_id, name) DO UPDATE SET name = EXCLUDED.name
	RETURNING id
), product_created AS (
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
), platform_cut_modifier AS (
	INSERT INTO earnings_modifiers
		(
			name,
			type,
			value,
			product_id
		)
	SELECT
		'platform_cut',
		'percentage',
		5.00,
		id
	FROM product_created
)
SELECT id FROM product_created
`

	// listing
	ProductGetByShop = `
SELECT * FROM product_catalog_view WHERE product_shop_id = $1
`
)
