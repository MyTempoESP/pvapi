package query

const (
	ListSaquesByShopQuery = `
SELECT
	sq.id,
	sq.description,
	sq.asked_value,
	sq.bank_name,
	sq.pix_cpf,
	sq.pix_titular,
	sq.transfer_details,
	sq.status,
	sq.created_at,
	u.name AS user_name,
	sh.name AS shop_name
FROM saques sq
JOIN users u ON u.id = sq.user_id
JOIN shops sh ON sh.id = sq.shop_id
WHERE sq.shop_id = $1
ORDER BY sq.created_at DESC
`
)
