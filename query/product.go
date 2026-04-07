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
		uniqueid,
		platform_cut,
		platform_cut_type
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
		uuidv4(),
		5.00,
		'percentage'
	)
RETURNING id
`

	ProductAddToConsignmentQuery = `
WITH inventory_check AS (
	SELECT product_decrement_inventory(@product_id, @quantity)
)
INSERT INTO packages
	(
		product_id,
		consignment_id,
		quantity,
		price,
		attributesValues,
		discount,
		featured
	)
SELECT
	@product_id,
	@consignment_id,
	@quantity,
	@price,
	@attr,
	@discount,
	@featured
FROM inventory_check
RETURNING id
`

	ProductTakeFromStock = `
SELECT product_decrement_inventory($1, $2)
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
	product_platform_cut,
	product_platform_cut_type,

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
	product_platform_cut,
	product_platform_cut_type,

	product_main_image,
	product_spec_images,
	product_desc_images,
	product_pack_images
FROM product_catalog_view
WHERE product_shop_id = $1
`

	ProductModerate = `
UPDATE products
SET moderation_status = @status
WHERE id = @product_id
`

	ProductPublicIndexByShopQuery = `
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
	product_platform_cut,
	product_platform_cut_type,

	product_main_image,
	product_spec_images,
	product_desc_images,
	product_pack_images
FROM product_catalog_view
WHERE product_shop_id = $1 AND product_moderation_status = 'approved'
`
)
