package query

const (
	CreateSpecificationQuery = `
INSERT INTO specifications (name, value, product_id) VALUES ($1, $2, $3)
`
)
