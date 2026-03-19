package query

const (
	CreateImageQuery = `
INSERT INTO images (filename, alt, tag, product_id) VALUES ($1, $2, $3, $4)
`
)
