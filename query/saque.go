package query

const (
	CreateSaqueQuery = `
INSERT INTO saques
	(shop_id, user_id, description, asked_value, bank_account_id)
VALUES
	($1, $2, $3, $4, $5)
RETURNING id
`

	ListSaquesByShopQuery = `
SELECT
	sq.id,
	sq.description,
	sq.asked_value,
	sq.bank_account_id,
	ba.bank_name AS bank_name,
	ba.pix_key,
	ba.pix_titular,
	ba.transfer_details,
	sq.status,
	sq.created_at,
	u.name AS user_name,
	sh.name AS shop_name
FROM saques sq
JOIN users u ON u.id = sq.user_id
JOIN shops sh ON sh.id = sq.shop_id
LEFT JOIN bank_accounts ba ON ba.id = sq.bank_account_id
WHERE sq.shop_id = $1
ORDER BY sq.created_at DESC
`

	ListAllSaquesQuery = `
SELECT
	sq.id,
	sq.description,
	sq.asked_value,
	sq.bank_account_id,
	ba.bank_name AS bank_name,
	ba.pix_key,
	ba.pix_titular,
	ba.transfer_details,
	sq.status,
	sq.created_at,
	u.name AS user_name,
	sh.name AS shop_name
FROM saques sq
JOIN users u ON u.id = sq.user_id
JOIN shops sh ON sh.id = sq.shop_id
LEFT JOIN bank_accounts ba ON ba.id = sq.bank_account_id
ORDER BY sq.created_at DESC
`

	FindSaqueByIDQuery = `
SELECT
	sq.id,
	sq.description,
	sq.asked_value,
	sq.bank_account_id,
	ba.bank_name AS bank_name,
	ba.pix_key,
	ba.pix_titular,
	ba.transfer_details,
	sq.status,
	sq.created_at,
	u.name AS user_name,
	sh.name AS shop_name
FROM saques sq
JOIN users u ON u.id = sq.user_id
JOIN shops sh ON sh.id = sq.shop_id
LEFT JOIN bank_accounts ba ON ba.id = sq.bank_account_id
WHERE sq.id = $1
`

	UpdateSaqueStatusQuery = `
UPDATE saques SET status = $2 WHERE id = $1
`

	DeleteSaqueQuery = `
DELETE FROM saques WHERE id = $1
`
)
