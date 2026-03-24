CREATE OR REPLACE VIEW public.consignees_info_full_view AS
SELECT
	c.id AS consignee_id,
	c.telephone AS phone,
	c.shop_id AS consignee_shop_id,

	u.name AS user_name,
	u.email AS user_email,

	a.street AS addrstreet,
	a.number AS addrnumber,
	a.neighborhood AS addrneigh,
	a.city AS addrcity,
	a.state AS addrstate,
	a.cep AS addrcep,
	a.complement AS addrcomplement
FROM public.consignees c
LEFT JOIN public.users u ON u.id = c.user_id
LEFT JOIN public.addresses a ON a.id = c.address_id;

CREATE OR REPLACE VIEW public.consignees_info_view AS
SELECT
	consignee_id,
	phone,
	-- ommitted shop_id
	user_name,
	user_email,
	addrstreet,
	addrnumber,
	addrneigh,
	addrcity,
	addrstate,
	addrcep,
	addrcomplement
FROM public.consignees_info_full_view;

