package query

const (
	CreateReportQuery = `
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

	FindReportQuery = `
SELECT
    *
FROM reports
WHERE id = $1
`

	DeleteReportQuery = `
DELETE FROM reports
WHERE id = $1
`
)
