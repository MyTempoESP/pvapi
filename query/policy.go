package query

const (
	CreatePolicyQuery = `
INSERT INTO policies
	(name, enabled, product_id)
VALUES
	($1, $2, $3)
`
)
