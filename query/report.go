package query

const (
    CreateReportQuery = `
INSERT INTO reports
    (
        field
    )
VALUES
    (
        @field
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
