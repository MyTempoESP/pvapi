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

	ResolvePickupEmailDataQuery = `
SELECT
	pl.name AS pickup_name,
	p.name AS product_name,
	SUM(ci.quantity)::integer AS quantity,
	a.addrstreet,
	a.addrnumber,
	a.addrcity,
	a.addrstate,
	a.addrcep,
	a.addrneigh,
	a.addrcomplement
FROM (
	SELECT unnest($1::integer[]) AS package_id,
	       unnest($2::integer[]) AS quantity
) ci
JOIN packages pkg ON pkg.id = ci.package_id
JOIN products p ON p.id = pkg.product_id
JOIN consignments con ON con.id = pkg.consignment_id
JOIN pickup_locations pl ON pl.id = con.pickup_location_id
LEFT JOIN aliased_addresses_view a ON a.address_id = pl.address_id
GROUP BY pl.id, pl.name, p.id, p.name,
	a.addrstreet, a.addrnumber, a.addrcity, a.addrstate,
	a.addrcep, a.addrneigh, a.addrcomplement
ORDER BY pl.name, p.name
`

	InsertCheckoutShopQuery = `
INSERT INTO checkout_shop
	(checkout_id, shop_id, checkout_unique_id, shop_value, shop_tax, shop_commission)
SELECT
	$1,
	con.shop_id,
	$2::uuid,
	SUM(pkg.price * pr.quantity),
	SUM((p.platform_cut / 100.0) * pkg.price * pr.quantity),
	SUM(CASE
		WHEN con.comission_type = 'PERCENT'
			THEN (con.comission_value / 100.0) * pkg.price * pr.quantity
		ELSE con.comission_value + (pkg.price * pr.quantity)
	END)
FROM package_reserved pr
JOIN packages pkg ON pkg.id = pr.package_id
JOIN products p ON p.id = pkg.product_id
JOIN consignments con ON con.id = pkg.consignment_id
WHERE pr.checkout_id = $1
GROUP BY con.shop_id
`

	UpdateCheckoutDiscrepancyQuery = `
UPDATE checkouts chk
SET value_discrepancy = chk.value - COALESCE(
	(SELECT SUM(cs.shop_value)
	 FROM checkout_shop cs
	 WHERE cs.checkout_id = chk.id),
	0
)
WHERE chk.id = $1
`

	ShopCheckoutsQuery = `
SELECT
	chk.id              AS checkout_id,
	chk.unique_id::text AS checkout_unique_id,
	chk.description     AS checkout_description,
	cs.shop_value       AS checkout_value,
	cs.shop_tax         AS checkout_tax,
	cs.shop_commission  AS checkout_commission
FROM checkouts chk
JOIN checkout_shop cs ON cs.checkout_id = chk.id
WHERE cs.shop_id = $1
ORDER BY chk.processed_at DESC
`
)
