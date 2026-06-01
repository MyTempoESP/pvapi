package query

const (
	// SaqueTotalsByShopQuery returns ASKED_VALUE sums grouped by status
	// for a given shop. Used by AvailableBalance.
	SaqueTotalsByShopQuery = `
SELECT
	COALESCE(SUM(asked_value), 0) AS total,
	status
FROM saques
WHERE shop_id = $1
  AND status IN ('pending', 'approved')
GROUP BY status
`

	// CheckoutEarningsByShopQuery returns the total earnings
	// (value - tax - commission) from confirmed checkouts for a given shop.
	CheckoutEarningsByShopQuery = `
SELECT COALESCE(SUM(cs.shop_value - cs.shop_tax - cs.shop_commission), 0) AS total
FROM checkout_shop cs
WHERE cs.shop_id = $1
`

	// LockShopRowQuery locks the shop row for the duration of a transaction
	// to serialize concurrent balance-mutating operations.
	LockShopRowQuery = `
SELECT id FROM shops WHERE id = $1 FOR UPDATE
`
)
