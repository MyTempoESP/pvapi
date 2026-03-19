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
SELECT * FROM get_shop_products(@shop_id, @desc_tag, @spec_tag, @pack_tag)
`
)
