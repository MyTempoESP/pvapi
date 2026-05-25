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
-- $3 arrives as integer (item.Quantity), but the tax/commission
-- arithmetic requires numeric. The CTE coerces $3 to numeric
-- once so it can be used in both the integer quantity column
-- (via ::integer cast) and the numeric multiplications below.
WITH params(qty) AS (VALUES ($3::numeric))
INSERT INTO package_reserved
	(
		package_id,
		checkout_id,
		quantity,
		tax,
		commission
	)
SELECT
	$1,
	$2,
	qty::integer,
	(p.platform_cut / 100.0) * pkg.price * qty,
	CASE
		-- matches consignmentmodel.ComissionPercent / ComissionFixed
		WHEN c.comission_type = 'PERCENT'
			THEN (c.comission_value / 100.0) * pkg.price * qty
		ELSE c.comission_value + (pkg.price * qty)
	END
FROM packages pkg
JOIN products p ON p.id = pkg.product_id
JOIN consignments c ON c.id = pkg.consignment_id
CROSS JOIN params
WHERE pkg.id = $1
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

	ShopCheckoutsQuery = `
SELECT
	chk.id              AS checkout_id,
	chk.unique_id::text AS checkout_unique_id,
	chk.description     AS checkout_description,
	chk.value           AS checkout_value,
	SUM(pr.tax)         AS checkout_tax,
	SUM(pr.commission)  AS checkout_commission
FROM checkouts chk
JOIN package_reserved pr ON pr.checkout_id = chk.id
JOIN packages pkg ON pkg.id = pr.package_id
JOIN consignments con ON con.id = pkg.consignment_id
WHERE con.shop_id = $1
GROUP BY chk.id
ORDER BY chk.processed_at DESC
`
)
