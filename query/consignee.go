package query

// Consignee = Monitor

const (
	CreateConsigneeQuery = `
INSERT INTO consignees
	(
		user_id, address_id,
		shop_id, telephone
	)
SELECT
	@user_id,
	insert_new_address(
		@addrstreet,
		@addrnumber,
		@addrneigh,
		@addrcity,
		@addrstate,
		@addrcep,
		@addrcomplement
	),
	@shop_id,
	@telephone
RETURNING id
`

	ListShopConsigneesQuery = `
SELECT
	v.consignee_id,
	v.phone,

	v.user_name,
	v.user_email,

	v.addrstreet,
	v.addrnumber,
	v.addrneigh,
	v.addrcity,
	v.addrstate,
	v.addrcep,
	v.addrcomplement
FROM consignees_info_full_view v
WHERE v.consignee_shop_id = $1
`
)
