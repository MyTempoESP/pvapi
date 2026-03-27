package query

import (
	"fmt"
	"maps"
	"strings"

	"github.com/jackc/pgx/v5"
)

// QueryBuilders are defined for situations where
// dynamic query building is needed, other than fixed
// output/input queries.

type Clause struct {
	Clauses   []string
	Prefix    string // SET/WHERE
	Separator string // ", "/" AND "/" OR "
	Params    pgx.NamedArgs
}

// Add gets a field name and a mapped value, associates
// the database atom with the specified value via a parameter
// mapping automatically fetched from the "field" name.
func (c *Clause) Add(field string, value any) {
	if value == nil {
		return
	}

	c.Clauses = append(c.Clauses,
		fmt.Sprintf("%s=%s", field, "@"+field))
	c.Params[field] = value
}

// AddMap should be used when a specific name is required for
// the mapping, for situations like "*.id" or COUNT, where
// it's not sensible to use the field name for the param.
func (c *Clause) AddMap(field, param string, value any) {
	if value == nil {
		return
	}

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
