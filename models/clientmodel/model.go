package clientmodel

import (
	"context"
	"fmt"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"

	sq "github.com/Masterminds/squirrel"
)

type Client struct {
	ID    int
	Field any
}

func (c Client) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	q := util.SQL

	fields := make(map[string]any)

	if c.Field != nil {
		fields["field"] = c.Field
	}

	if len(fields) == 0 {
		return
	}

	query, params, err := q.Update("clients").
		SetMap(fields).
		Where(sq.Eq{"id": c.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	_, err = base.Exec(ctx, query, params...)

	if err != nil {
		return fmt.Errorf(
			"could not update client, query: %q, %w", query, err)
	}

	return
}

func (c ClientBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	return model.Create(
		ctx,
		base,
		query.CreateClientQuery,
		util.Args{
			"field": c.Field,
		},
	)
}

func List(
	ctx context.Context,
	base db.Queryable,
) ([]Client, error) {
	return util.WithCollect[Client](
		ctx,
		base,
		query.ListAllClientsQuery,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (Client, error) {
	return util.WithCollectExactlyOneRow[Client](
		ctx,
		base,
		query.FindClientQuery,
		id,
	)
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteClientQuery, id)
	return err
}
