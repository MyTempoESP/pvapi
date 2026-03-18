package query

const (
	// creating
	ProductCreate = `
INSERT INTO products
	(
		name, price, stock,
		discount, featured, shop_id
	)
VALUES
	(
		@name, @price, @stock,
		@discount, @featured, @shop_id
	)
RETURNING id
`

	// listing
	ProductGetByShop = `
SELECT
	p.id       AS product_id,
	p.name     AS product_name,
	p.price    AS product_price,
	p.stock    AS product_stock,
	p.discount AS product_discount,
	p.featured AS product_featured,
	COALESCE(sp.listing, '[]'::json) AS product_specifications,
	COALESCE(pp.listing, '[]'::json) AS product_policies,
	COALESCE(em.listing, '[]'::json) AS product_earnings_modifiers
FROM products p
LEFT JOIN LATERAL (
	SELECT
		json_agg(
			json_build_object(
				'spec_name'  , s.name,
				'spec_value' , s.value,
				'spec_images', COALESCE(img.img_array, '[]'::json)
			)
		) AS listing
	FROM specifications s
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
	WHERE s.product_id = p.id
) sp ON true
LEFT JOIN LATERAL (
	SELECT
        json_agg(
            json_build_object(
                'pol_name'   , po.name,
                'pol_value'  , po.value,
                'pol_enabled', po.enabled
            )
        ) AS listing
    FROM product_policies po
    WHERE po.product_id = p.id
) pp ON true
LEFT JOIN LATERAL (
    SELECT
        json_agg(
            json_build_object(
                'emod_name' , e.name,
                'emod_type' , e.type,
                'emod_value', e.value
            )
        ) AS listing
    FROM earnings_modifiers e
    WHERE e.product_id = p.id
) em ON true
WHERE p.shop_id = $1
`
)
