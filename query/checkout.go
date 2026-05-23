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

	FindCheckoutByCartRefQuery = `
SELECT unique_id
FROM checkouts
WHERE cart_id = $1
  AND external_reference = $2
`

	CheckPackageAvailabilityQuery = `
SELECT
	pkg.id AS package_id,
	pkg.quantity AS total_allocated,
	COALESCE(
		(SELECT SUM(pr.quantity)
		 FROM package_reserved pr
		 WHERE pr.package_id = pkg.id),
		0
	) AS already_reserved,
	pkg.quantity - COALESCE(
		(SELECT SUM(pr.quantity)
		 FROM package_reserved pr
		 WHERE pr.package_id = pkg.id),
		0
	) AS available
FROM packages pkg
WHERE pkg.product_id = $1
  AND pkg.consignment_id = $2
FOR UPDATE OF pkg
`

	FinishCartQuery = `
UPDATE carts
SET session_id = 'FINISHED-' || $2
WHERE id = $1
`
)
