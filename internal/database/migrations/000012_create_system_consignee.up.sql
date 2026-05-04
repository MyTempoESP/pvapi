ALTER TABLE public.consignees
ALTER COLUMN shop_id DROP NOT NULL;

WITH inserted_user AS (
	INSERT INTO users
	(
		name,
		email,
		password_hash
	)
	VALUES
	(
		'system',
		'sistema@pvauto.online',
		'no-login'
	)
	RETURNING id
)
INSERT INTO consignees
(
	user_id, address_id,
	shop_id, telephone
)
SELECT
	id,
	insert_new_address(
		'dummy',
		0,
		'dummy',
		'dummy',
		'dm',
		'00000-00',
		''
	),
	NULL,
	''
FROM inserted_user
RETURNING id;

