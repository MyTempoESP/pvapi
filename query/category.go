package query

const (
	CreateCategoryQuery = `
INSERT INTO categories (name, description, shop_id) VALUES ($1, $2, $3)
RETURNING id
`

	CategoryListByShopQuery = `
SELECT * FROM categories
WHERE shop_id = $1
`
)
