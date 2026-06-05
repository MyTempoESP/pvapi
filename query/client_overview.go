package query

const (
	ClientOverviewCountsQuery = `
SELECT
	COUNT(DISTINCT pr.checkout_id) AS total_orders,
	COUNT(*) FILTER (WHERE pr.status != 'recebido') AS pending_count
FROM package_reserved pr
WHERE pr.client_id = $1
`

	ClientOverviewPackagesQuery = `
SELECT
	chk.unique_id::text   AS checkout_unique_id,
	pr.reserved_at,
	pr.status,
	pr.shop_value,
	pr.quantity,

	pick.pickup_id,
	pick.pickup_name,
	pick.pickup_datetime,

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
LEFT JOIN LATERAL (
	SELECT pl2.id AS pickup_id, pl2.name AS pickup_name, pl2.pickup_at AS pickup_datetime, pl2.address_id AS pickup_address_id
	FROM consignment_pickup_locations cpl
	JOIN pickup_locations pl2 ON pl2.id = cpl.pickup_location_id
	WHERE cpl.consignment_id = con.id
	LIMIT 1
) pick ON true
LEFT JOIN aliased_addresses_view addr ON addr.address_id = pick.pickup_address_id
WHERE pr.client_id = $1
ORDER BY pr.reserved_at DESC
`
)
