package reportmodel

import (
	"context"
	"fmt"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"

	sq "github.com/Masterminds/squirrel"
)

type Report struct {
	ID    int
	Field any
}

func (r Report) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	q := util.SQL

	fields := make(map[string]any)

	if len(fields) == 0 {
		return
	}

	update := q.Update("consignments").
		SetMap(fields).
		Where(sq.Eq{"id": r.ID})

	query, params, err := update.ToSql()

	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	_, err = base.Exec(ctx, query, params...)

	if err != nil {
		return fmt.Errorf(
			"could not update consignment, query: %q, %w", query, err)
	}

	return err
}

func (r ReportBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	return model.Create(
		ctx,
		base,
		query.CreateReportQuery,
		util.Args{
			"field": r.Field,
		},
	)
}

func List(
	ctx context.Context,
	base db.Queryable,
) ([]Report, error) {
	return util.WithCollect[Report](
		ctx,
		base,
		query.ListAllReportsQuery,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (Report, error) {
	return util.WithCollectExactlyOneRow[Report](
		ctx,
		base,
		query.FindReportQuery,
		id,
	)
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteReportQuery, id)
	return err
}
