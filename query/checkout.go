package query

const (
	InsertCheckoutQuery = `
INSERT INTO checkouts
	(
		status,
		description,
		api_client_id,
		client_id,
		callback
	)
VALUES
	(
		'paid',
		@description,
		@api_client_id,
		@client_id,
		@callback
	)
RETURNING id, unique_id
`

	InsertPackageReservedQuery = `
WITH decrement AS (
	UPDATE packages
	SET quantity = quantity - $3
	WHERE id = $1 AND quantity >= $3
	RETURNING id, product_id, consignment_id, price, discount
)
INSERT INTO package_reserved
	(package_id, checkout_id, quantity, tax, commission, shop_id, consignment_id, shop_value, status, pickup_location_id)
SELECT
	d.id,
	$2,
	$3::integer,
	(p.platform_cut / 100.0) * d.price * (1.0 - d.discount / 100.0) * $3::numeric,
	CASE
		-- FIXED: commission is the flat comission_value per checkout,
		--        not added to the package price
		WHEN c.comission_type = 'PERCENT'
			THEN (c.comission_value / 100.0) * d.price * (1.0 - d.discount / 100.0) * $3::numeric
		ELSE c.comission_value
	END,
	c.shop_id,
	d.consignment_id,
	d.price * (1.0 - d.discount / 100.0) * $3::numeric,
	'pending',
	$4
FROM decrement d
JOIN products p ON p.id = d.product_id
JOIN consignments c ON c.id = d.consignment_id
RETURNING unique_id, package_id
`

	CheckRateLimitQuery = `
SELECT EXTRACT(EPOCH FROM (NOW() - chk.processed_at))::bigint AS seconds_ago
FROM checkouts chk
WHERE chk.client_id = $1
ORDER BY chk.processed_at DESC
LIMIT 1
`

	CheckPackageAvailabilityQuery = `
SELECT quantity AS available
FROM packages
WHERE id = $1
FOR UPDATE
`

	FindCheckoutIDByUniqueIDQuery = `
SELECT id FROM checkouts WHERE unique_id = $1
`

	FindPackageReservedByCheckoutQuery = `
SELECT unique_id, package_id
FROM package_reserved
WHERE checkout_id = $1
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
JOIN consignment_pickup_locations cpl ON cpl.consignment_id = con.id
JOIN pickup_locations pl ON pl.id = cpl.pickup_location_id
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

	UpdateCheckoutClientIDQuery = `
UPDATE checkouts
SET client_id = $1
WHERE id = $2
`

	UpdatePackageReservedClientIDQuery = `
UPDATE package_reserved
SET client_id = $1
WHERE checkout_id = $2
`

	UpdateCheckoutShopClientIDByCheckoutIDQuery = `
UPDATE checkout_shop
SET client_id = $1
WHERE checkout_id = $2
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

	ValidatePickupLocationQuery = `
SELECT EXISTS (
	SELECT 1 FROM packages pkg
	JOIN consignment_pickup_locations cpl ON cpl.consignment_id = pkg.consignment_id
	WHERE pkg.id = $1 AND cpl.pickup_location_id = $2
)
`

	EnrichedQRCodeEntriesQuery = `
SELECT
	pr.unique_id::text,
	p.id            AS product_id,
	pr.quantity,
	pkg.price,
	pkg.discount,
	pr.pickup_location_id,
	pl.name         AS pickup_name,
	a.addrstreet,
	a.addrnumber::text,
	a.addrcity,
	a.addrstate,
	a.addrcep,
	a.addrneigh,
	a.addrcomplement
FROM package_reserved pr
JOIN packages pkg ON pkg.id = pr.package_id
JOIN products p ON p.id = pkg.product_id
JOIN pickup_locations pl ON pl.id = pr.pickup_location_id
LEFT JOIN aliased_addresses_view a ON a.address_id = pl.address_id
WHERE pr.checkout_id = $1
ORDER BY pr.id
`
)
