package query

const (
	CheckClientExistsQuery = `SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)`

	ClientSalesQuery = `
SELECT
	cs.checkout_id,
	cs.shop_id,
	cs.client_id,
	cs.shop_value,
	cs.created_at,
	pr.quantity,
	pr.status           AS package_status,
	COALESCE(prod.name, '') AS product_name
FROM checkout_shop cs
JOIN package_reserved pr ON pr.checkout_id = cs.checkout_id AND pr.shop_id = cs.shop_id
LEFT JOIN packages pkg ON pkg.id = pr.package_id
LEFT JOIN products prod ON prod.id = pkg.product_id
WHERE cs.client_id = $1 AND cs.shop_id = $2
ORDER BY cs.created_at DESC, pr.id
`
)
