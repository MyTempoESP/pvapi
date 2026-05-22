---
to: query/<%= h.changeCase.no(name) %>.go
---
<%
    Flat = h.changeCase.no(name)
    Table = h.inflection.pluralize(Flat)

    Class = h.changeCase.pascal(name)
    Model = Flat + "model"

    Char = name.charAt(0)
%>package query

const (
    Create<%= Class %>Query = `
INSERT INTO <%= Table %>
    (
        field
    )
VALUES
    (
        @field
    )
RETURNING id
`

    ListAll<%= h.inflection.pluralize(Class) %>Query = `
SELECT
    *
FROM <%= Table %>
`

    Find<%= Class %>Query = `
SELECT
    *
FROM <%= Table %>
WHERE id = $1
`

    Delete<%= Class %>Query = `
DELETE FROM <%= Table %>
WHERE id = $1
`
)
