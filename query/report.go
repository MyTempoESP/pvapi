package query

const (
	CreateReportQuery = `
-- TODO: resolving consignee by u.name = @consignee_name is fragile;
-- replace with consignee ID once the frontend passes it.
WITH consignee AS (
	SELECT c.*
	FROM consignees c
	LEFT JOIN users u ON c.user_id = u.id
	WHERE u.name = @consignee_name
)
INSERT INTO reports
	(
		description,
		moment,
		consignee_id,
		scope,
		consignment_id,
		priority,
		status,
		type
	)
VALUES
	(
		@description,
		@moment,
		(SELECT id FROM consignee),
		@scope,
		@consignment_id,
		@priority,
		@status,
		@type
	)
RETURNING id
`

	ListAllReportsQuery = `
SELECT r.*, u.name AS consignee_name
FROM reports r
LEFT JOIN consignees c ON r.consignee_id = c.id
LEFT JOIN users u ON c.user_id = u.id
`

	ListAllReportsByShopQuery = `
SELECT r.*, u.name AS consignee_name,
	COALESCE(p.name, c.slug) AS order_name
FROM shops s
JOIN reports r ON
	(r.consignment_id IN (SELECT c.id FROM consignments c WHERE c.shop_id = s.id))
	OR
	(r.product_id IN (SELECT p.id FROM products p WHERE p.shop_id = s.id))
LEFT JOIN consignees co ON r.consignee_id = co.id
LEFT JOIN users u ON co.user_id = u.id
LEFT JOIN products p ON r.product_id = p.id
LEFT JOIN consignments c ON r.consignment_id = c.id
WHERE s.id = $1
`

	ListAllReportsByConsigneeQuery = `
SELECT r.*, u.name AS consignee_name,
	COALESCE(p.name, co.slug) AS order_name
FROM consignees c
JOIN reports r ON r.consignee_id = c.id
LEFT JOIN users u ON c.user_id = u.id
LEFT JOIN products p ON r.product_id = p.id
LEFT JOIN consignments co ON r.consignment_id = co.id
WHERE c.id = $1
`

	FindReportQuery = `
SELECT r.*, u.name AS consignee_name,
	COALESCE(p.name, co.slug) AS order_name
FROM reports r
LEFT JOIN consignees c ON r.consignee_id = c.id
LEFT JOIN users u ON c.user_id = u.id
LEFT JOIN products p ON r.product_id = p.id
LEFT JOIN consignments co ON r.consignment_id = co.id
WHERE r.id = $1
`

	CreateProductReportQuery = `
-- TODO: resolving consignee by u.name = @consignee_name is fragile;
-- replace with consignee ID once the frontend passes it.
WITH consignee AS (
	SELECT c.*
	FROM consignees c
	LEFT JOIN users u ON c.user_id = u.id
	WHERE u.name = @consignee_name
)
INSERT INTO reports
	(
		description,
		moment,
		consignee_id,
		scope,
		product_id,
		priority,
		status,
		type
	)
VALUES
	(
		@description,
		@moment,
		(SELECT id FROM consignee),
		@scope,
		@product_id,
		@priority,
		@status,
		@type
	)
RETURNING id
`

	DeleteReportQuery = `
DELETE FROM reports
WHERE id = $1
`

	CreateReportWithConsigneeIDQuery = `
INSERT INTO reports
	(description, moment, consignee_id, scope, consignment_id, priority, status, type)
VALUES
	(@description, @moment, @consignee_id, @scope, @consignment_id, @priority, @status, @type)
RETURNING id
`

	CreateProductReportWithConsigneeIDQuery = `
INSERT INTO reports
	(description, moment, consignee_id, scope, product_id, priority, status, type)
VALUES
	(@description, @moment, @consignee_id, @scope, @product_id, @priority, @status, @type)
RETURNING id
`

	FindReportScopeQuery = `
SELECT scope FROM reports WHERE id = $1
`
)
