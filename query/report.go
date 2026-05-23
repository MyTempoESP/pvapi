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
SELECT
    *
FROM reports
`

	ListAllReportsByShopQuery = `
SELECT r.*
FROM shops s
JOIN reports r ON
	(r.consignment_id IN (SELECT c.id FROM consignments c WHERE c.shop_id = s.id))
	OR
	(r.product_id IN (SELECT p.id FROM products p WHERE p.shop_id = s.id))
WHERE s.id = $1
`

	ListAllReportsByConsigneeQuery = `
SELECT
    r.*
FROM consignees c
JOIN reports r ON r.consignee_id = c.id
WHERE c.id = $1
`

	FindReportQuery = `
SELECT
    *
FROM reports
WHERE id = $1
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

	FindReportScopeQuery = `
SELECT scope FROM reports WHERE id = $1
`
)
