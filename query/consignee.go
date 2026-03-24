package query

// Consignee = Monitor

const (
	CreateConsigneeQuery = `
WITH address AS (
	INSERT INTO addresses
		(
			street,
			number,
			neighborhood,
			city,
			state,
			cep,
			complement
		)
	VALUES
		(
			@addrstreet, @addrnumber,
			@addrneigh, @addrcity,
			@addrstate, @addrcep,
			@addrcomplement
		)
	RETURNING id
)
INSERT INTO consignees
	(
		user_id, address_id,
		shop_id, telephone
	)
SELECT
	@user_id, a.id, @shop_id, @telephone
FROM address a
`
)
