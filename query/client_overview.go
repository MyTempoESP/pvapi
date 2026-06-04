package query

const (
	ClientOverviewCountsQuery = `
SELECT
	COUNT(DISTINCT pr.checkout_id) AS total_orders,
	COUNT(*) FILTER (WHERE pr.status != 'recebido') AS pending_count
FROM package_reserved pr
JOIN checkouts chk ON chk.id = pr.checkout_id
WHERE chk.client_id = $1
`

	ClientOverviewPackagesQuery = `
SELECT
	chk.unique_id::text   AS checkout_unique_id,
	pr.reserved_at,
	pr.status,
	pr.shop_value,
	pr.quantity,

	pl.id                 AS pickup_id,
	pl.name               AS pickup_name,
	pl.pickup_at          AS pickup_datetime,

	addr.addrstreet       AS addrstreet,
	addr.addrnumber       AS addrnumber,
	addr.addrneigh        AS addrneigh,
	addr.addrcity         AS addrcity,
	addr.addrstate        AS addrstate,
	addr.addrcep          AS addrcep,
	addr.addrcomplement   AS addrcomplement,

	prod.id               AS product_id,
	prod.name             AS product_name,
	prod.code             AS product_code
FROM package_reserved pr
JOIN checkouts chk ON chk.id = pr.checkout_id
LEFT JOIN packages pkg ON pkg.id = pr.package_id
LEFT JOIN products prod ON prod.id = pkg.product_id
LEFT JOIN consignments con ON con.id = pr.consignment_id
LEFT JOIN pickup_locations pl ON pl.id = con.pickup_location_id
LEFT JOIN aliased_addresses_view addr ON addr.address_id = pl.address_id
WHERE chk.client_id = $1
ORDER BY pr.reserved_at DESC
`
)
