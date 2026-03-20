package query

const (
	CreateCategoryQuery = `
INSERT INTO categories (name, description, shop_id) VALUES ($1, $2, $3)
`
)
