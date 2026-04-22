package query

const (
	// creating
	ProductCreate = `
WITH product_category_inserted AS (
	INSERT INTO categories
		(
			name,
			description,
			shop_id
		)
	VALUES
		(
			@category,
			@category,
			@shop_id
		)
	ON CONFLICT (shop_id, name) DO NOTHING
	RETURNING id
), product_category AS (
	SELECT id FROM product_category_inserted
	UNION ALL
	SELECT id FROM categories
	WHERE name = @category AND shop_id = @shop_id
	LIMIT 1
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
SELECT
	@name,
	@price,
	@stock,
	@discount,
	@featured,
	@shop_id,
	'pending',
	c.id,
	@description,
	@attr,
	@code,
	uuidv4(),
	5.00,
	'percentage'
FROM product_category c
RETURNING id
`

	// TODO: only approved products can get into consignments
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
	product_shop_id,
	product_shop_name
FROM product_catalog_view
WHERE product_id = $1
`

	ProductPublicFindQuery = `
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
	product_shop_id,
	product_shop_name
FROM product_catalog_view
WHERE product_id = $1 AND product_moderation_status = 'approved'
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
	product_shop_id,
	product_shop_name
FROM product_catalog_view
WHERE product_shop_id = $1
`

	ProductModerate = `
UPDATE products
SET
	moderation_status = @status,
	platform_cut      = @platform_cut,
	platform_cut_type = @platform_cut_type
WHERE id = @id
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
	product_shop_id,
	product_shop_name
FROM product_catalog_view
WHERE product_shop_id = $1 AND product_moderation_status = 'approved'
`

	ProductsPendingModerationQuery = `
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
	product_shop_id,
	product_shop_name
FROM product_catalog_view
WHERE product_moderation_status = 'pending'
`

	ProductListInConsignmentQuery = `
SELECT
	c.product_id,
	c.product_name,
	c.product_price,
	c.product_stock,
	c.product_discount,
	c.product_featured,
	c.product_attr->'main_image' AS product_main_image,
	c.product_description,
	c.product_category_id,
	c.product_shop_id,
	c.product_code,
	c.product_uuid,

	p.price AS package_price,
	p.quantity AS package_quantity
FROM packages p
LEFT JOIN product_catalog_view c ON c.product_id = p.product_id
WHERE p.consignment_id = $1
`

	ProductGetAttrQuery = `
SELECT p.attr
FROM products p
WHERE p.id = $1
`

	ProductHasImage = `
SELECT
	(
		COALESCE(p.attr->'desc_images', '[]'::jsonb) ||
		COALESCE(p.attr->'spec_images', '[]'::jsonb) ||
		COALESCE(p.attr->'pack_images', '[]'::jsonb)
	) @> to_jsonb($2::text) AS hasImage
FROM products p
WHERE p.id = $1;
`

	ProductDeleteImageQuery = `
UPDATE products
SET attr = jsonb_set(
		jsonb_set(
			jsonb_set(
				attr,
				'{pack_images}',
				COALESCE(NULLIF(attr->'pack_images',
					'null'::jsonb), '[]'::jsonb) - $2::text
			),
			'{spec_images}',
			COALESCE(NULLIF(attr->'spec_images', 'null'::jsonb),
					'[]'::jsonb) - $2::text
		),
		'{desc_images}',
		COALESCE(NULLIF(attr->'desc_images', 'null'::jsonb),
				'[]'::jsonb) - $2::text
	)
WHERE id = $1;
`

	DeleteProduct = `
DELETE FROM products
WHERE id = $1
`
)
