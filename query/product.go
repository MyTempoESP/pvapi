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

	ProductAddToConsignmentQuery = `
INSERT INTO packages
	(
		product_id,
		consignment_id,
		quantity,
		price
	)
VALUES
	(
		@product_id,
		@consignment_id,
		@quantity,
		@price
	)
`

	ProductTakeFromStock = `
UPDATE products SET stock = stock - $2 WHERE id = $1
`

	FindProductQuery = `
SELECT
	product_id,
	product_name,
	product_price,
	product_stock,
	product_discount,
	product_featured,
	product_moderation_status,
	product_description,
	product_category,
	product_uuid,
	product_code,

	product_attr,

	product_specifications,

	product_p_cname,
	product_p_cmod,
	product_p_ccolor,

	product_platform_cut,

	product_main_image,
	product_spec_images,
	product_desc_images,
	product_pack_images
FROM product_catalog_view
WHERE product_id = $1
`

	// listing
	ProductGetByShop = `
SELECT
	product_id,
	product_name,
	product_price,
	product_stock,
	product_discount,
	product_featured,
	product_moderation_status,
	product_description,
	product_category,
	product_uuid,
	product_code,

	product_attr,

	product_specifications,

	product_p_cname,
	product_p_cmod,
	product_p_ccolor,

	product_platform_cut,

	product_main_image,
	product_spec_images,
	product_desc_images,
	product_pack_images
FROM product_catalog_view
WHERE product_shop_id = $1
`
)
