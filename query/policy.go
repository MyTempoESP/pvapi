package query

const (
	CreatePolicyQuery = `
INSERT INTO product_policies
	(name, value, enabled, product_id)
VALUES
	($1, $2, $3, $4)
`
)
