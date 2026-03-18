# Product queries

## ProductGetByShop

Fetches products belonging to a specific shop_id

```sql
SELECT
	p.id       AS product_id,
	p.name     AS product_name,
	p.price    AS product_price,
	p.stock    AS product_stock,
	p.discount AS product_discount,
	p.featured AS product_featured,
	COALESCE(sp.listing, '[]'::json) AS product_specifications
FROM products p
LEFT JOIN LATERAL (
	SELECT
        -- Aggregate ("collect") all specifications from a product
		json_agg(
			json_build_object(
				'spec_name'  , s.name,
				'spec_value' , s.value,
				'spec_images', COALESCE(img.img_array, '[]'::json)
			)
		) AS listing
	FROM specifications s

	-- Get all the images referencing this specification
	LEFT JOIN LATERAL (
		SELECT
			json_agg(
				json_build_object(
					'filename', i.filename,
					'alt'     , i.alt
				)
			) AS img_array
		FROM specification_images i
		WHERE i.specification_id = s.id
	) img ON true

	-- Filter by product id
	WHERE s.product_id = p.id
) sp ON true
WHERE p.shop_id = $1
```
