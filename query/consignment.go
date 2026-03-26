package query

const (
	CreateConsignmentQuery = `
INSERT INTO consignments
    (
        field
    )
VALUES
    (
        @field
    )
RETURNING id
`

	ListAllConsignmentsQuery = `
SELECT
    *
FROM consignments
`

	FindConsignmentQuery = `
SELECT
    *
FROM consignments
WHERE id = $1
`

	DeleteConsignmentQuery = `
DELETE FROM consignments
WHERE id = $1
`
)
