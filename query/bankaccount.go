package query

const (
	CreateBankAccountQuery = `
INSERT INTO bank_accounts
	(shop_id, name, bank_name, pix_key, pix_titular, transfer_details)
VALUES
	($1, $2, $3, $4, $5, $6)
RETURNING id
`

	ListBankAccountsByShopQuery = `
SELECT
	ba.id,
	ba.shop_id,
	ba.name,
	ba.bank_name,
	ba.pix_key,
	ba.pix_titular,
	ba.transfer_details,
	ba.created_at
FROM bank_accounts ba
WHERE ba.shop_id = $1
ORDER BY ba.created_at DESC
`

	FindBankAccountByIDQuery = `
SELECT
	ba.id,
	ba.shop_id,
	ba.name,
	ba.bank_name,
	ba.pix_key,
	ba.pix_titular,
	ba.transfer_details,
	ba.created_at
FROM bank_accounts ba
WHERE ba.id = $1
`

	UpdateBankAccountQuery = `
UPDATE bank_accounts
SET name = $2,
    bank_name = $3,
    pix_key = $4,
    pix_titular = $5,
    transfer_details = $6
WHERE id = $1
`

	DeleteBankAccountQuery = `
DELETE FROM bank_accounts WHERE id = $1
`

	ShopHasBankAccountQuery = `
SELECT (ba.id IS NOT NULL) AS hasBankAccount
FROM shops s
LEFT JOIN bank_accounts ba ON ba.shop_id = s.id AND ba.id = $2
WHERE s.id = $1
`
)
