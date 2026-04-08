package query

const (
	JSONBCatFormat = `
jsonb_set(%s, '{%s}', COALESCE(attr->'%[2]s', '[]'::jsonb) || ?::jsonb)
`

	JSONBSetFormat = `
jsonb_set(%s, '{%s}', ?::jsonb)
`
)
