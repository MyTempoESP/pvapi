package query

const (
	InsertCheckoutQuery = `
INSERT INTO checkouts
	(
		cart_id,
		value,
		status,
		description,
		external_reference
	)
VALUES
	(
		@cart_id,
		@value,
		'paid',
		@description,
		@external_reference
	)
RETURNING unique_id
`

	InsertPackageReservedQuery = `
INSERT INTO package_reserved
	(
		package_id,
		checkout_id,
		quantity
	)
VALUES
	(
		$1,
		$2,
		$3
	)
`

	FindCheckoutBySessionRefQuery = `
SELECT unique_id
FROM checkouts
JOIN carts ON carts.id = checkouts.cart_id
WHERE carts.session_id IN ($1, 'FINISHED-' || $2)
  AND external_reference = $2
`

	CheckPackageAvailabilityQuery = `
SELECT
	pkg.quantity - COALESCE(
		(SELECT SUM(pr.quantity)
		 FROM package_reserved pr
		 WHERE pr.package_id = pkg.id),
		0
	) AS available
FROM packages pkg
WHERE pkg.id = $1
FOR UPDATE OF pkg
`

	FinishCartQuery = `
UPDATE carts
SET session_id = 'FINISHED-' || $2
WHERE session_id = $1
`

	FindCartBySessionQuery = `
SELECT id FROM carts
WHERE session_id = $1
  AND session_id NOT LIKE 'FINISHED-%'
`

	CountCartItemsQuery = `
SELECT COUNT(*) FROM cart_product WHERE cart_id = $1
`

	FindCheckoutIDByUniqueIDQuery = `
SELECT id FROM checkouts WHERE unique_id = $1
`
)
