package query

import (
	"fmt"
	"maps"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Clause struct {
	Clauses   []string
	Prefix    string // SET/WHERE
	Separator string // ", "/" AND "/" OR "
	Params    pgx.NamedArgs
}

func (c *Clause) Add(field, param string, value any) {
	c.Clauses = append(c.Clauses,
		fmt.Sprintf("%s=%s", field, "@"+param))
	c.Params[param] = value
}

// Build makes the clause into a full form,
// it also drops if there is less than 1 clause
// defined.
func (c *Clause) Build() string {
	if len(c.Clauses) > 0 {
		return fmt.Sprintf(
			"%s %s", c.Prefix,
			strings.Join(c.Clauses, c.Separator),
		)
	}
	return ""
}

type Query interface {
	Build() (string, pgx.NamedArgs)
}

type UpdateQuery struct {
	Set   *Clause
	Where *Clause
	Model string
}

func Update(model string) UpdateQuery {
	return UpdateQuery{
		Set: &Clause{
			Prefix:    "SET",
			Separator: ", ",
			Clauses:   make([]string, 0),
			Params:    pgx.NamedArgs{},
		},

		Where: &Clause{
			Prefix:    "WHERE",
			Separator: " AND ",
			Clauses:   make([]string, 0),
			Params:    pgx.NamedArgs{},
		},

		Model: model,
	}
}

func (update *UpdateQuery) Build() (string, pgx.NamedArgs) {
	params := pgx.NamedArgs{}

	maps.Copy(params, update.Set.Params)
	maps.Copy(params, update.Where.Params)

	query := fmt.Sprintf("UPDATE %s %s %s",
		update.Model,
		update.Set.Build(),
		update.Where.Build(),
	)

	return query, params
}
