---
to: models/<%= h.changeCase.no(name) %>model/model.go
---
<%
    Flat = h.changeCase.no(name)
    Table = h.inflection.pluralize(Flat)

    Class = h.changeCase.pascal(name)
    Model = Flat + "model"

    Char = name.charAt(0)
%>package <%= Model %>

import (
	"context"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/util"
)

type <%= Class %> struct {
    ID    int
    Field any
}

func (<%= Char %> <%= Class %>) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	update := query.Update("<%= Table %> AS <%= Flat %>")

	update.Set.Add("field", <%= Char %>.Field)

	update.Where.AddMap("<%= Flat %>.id", "<%= Flat %>_id", <%= Char %>.ID)

	query, params := update.Build()

    _, err = base.Exec(ctx, query, params)

    return err
}

func (<%= Char %> <%= Class %>Blueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	return model.Create(
        ctx,
        base,
        query.Create<%= Class %>Query,
        util.Args{
            "field": <%= Char %>.Field,
        },
	)
}

func List(
	ctx context.Context,
	base db.Queryable,
) ([]<%= Class %>, error) {
	return util.WithCollect[<%= Class %>](
		ctx,
		base,
		query.ListAll<%= h.inflection.pluralize(Class) %>Query,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (<%= Class %>, error) {
	return util.WithCollectExactlyOneRow[<%= Class %>](
		ctx,
		base,
		query.Find<%= Class %>Query,
		id,
	)
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.Delete<%= Class %>Query, id)
	return err
}
