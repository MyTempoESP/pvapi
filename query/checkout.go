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
WITH decrement AS (
	UPDATE packages
	SET quantity = quantity - $3
	WHERE id = $1 AND quantity >= $3
	RETURNING id, product_id, consignment_id, price
)
INSERT INTO package_reserved
	(package_id, checkout_id, quantity, tax, commission, shop_id, consignment_id, shop_value, status)
SELECT
	d.id,
	$2,
	$3::integer,
	(p.platform_cut / 100.0) * d.price * $3::numeric,
	CASE
		-- FIXED: commission is the flat comission_value per checkout,
		--        not added to the package price
		WHEN c.comission_type = 'PERCENT'
			THEN (c.comission_value / 100.0) * d.price * $3::numeric
		ELSE c.comission_value
	END,
	c.shop_id,
	d.consignment_id,
	d.price * $3::numeric,
	'pending'
FROM decrement d
JOIN products p ON p.id = d.product_id
JOIN consignments c ON c.id = d.consignment_id
`

	FindCheckoutBySessionRefQuery = `
SELECT unique_id
FROM checkouts
JOIN carts ON carts.id = checkouts.cart_id
WHERE carts.session_id IN ($1, 'FINISHED-' || $2)
  AND external_reference = $2
`

	CheckPackageAvailabilityQuery = `
SELECT quantity AS available
FROM packages
WHERE id = $1
FOR UPDATE
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
	pr.shop_id,
	$2::uuid,
	SUM(pr.shop_value),
	SUM(pr.tax),
	SUM(pr.commission)
FROM package_reserved pr
WHERE pr.checkout_id = $1
GROUP BY pr.shop_id
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
