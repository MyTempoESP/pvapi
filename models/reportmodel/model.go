package reportmodel

import (
	"context"
	"fmt"
	"time"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"

	sq "github.com/Masterminds/squirrel"
)

type Report struct {
	ID            int       `db:"id"`
	Description   string    `db:"description"`
	Moment        string    `db:"moment"`
	ConsigneeID   int       `db:"consignee_id"`
	ConsigneeName string    `db:"consignee_name"`
	Scope         string    `db:"scope"`
	ConsignmentID *int      `db:"consignment_id"`
	ProductID     *int      `db:"product_id"`
	Priority      string    `db:"priority"`
	Status        string    `db:"status"`
	Type          string    `db:"type"`
	CreatedAt     time.Time `db:"created_at"`
}

func (r Report) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	if true {
		return util.NotImplemented("Atualizar ocorrencia")
	}

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
	if r.Scope == "product" {
		return model.Create(
			ctx,
			base,
			query.CreateProductReportQuery,
			util.Args{
				"description":    r.Description,
				"moment":         r.Moment,
				"consignee_name": r.Monitor,
				"scope":          r.Scope,
				"product_id":     r.OrderID,
				"priority":       r.Priority,
				"status":         r.Status,
				"type":           r.Type,
			},
		)
	}

	return model.Create(
		ctx,
		base,
		query.CreateReportQuery,
		util.Args{
			"description":    r.Description,
			"moment":         r.Moment,
			"consignee_name": r.Monitor,
			"scope":          r.Scope,
			"consignment_id": r.OrderID,
			"priority":       r.Priority,
			"status":         r.Status,
			"type":           r.Type,
		},
	)
}

func ListByConsignee(
	ctx context.Context,
	base db.Queryable,
	consignee int,
) ([]Report, error) {
	return util.WithCollect[Report](
		ctx,
		base,
		query.ListAllReportsByConsigneeQuery,
		consignee,
	)
}

func ListByShop(
	ctx context.Context,
	base db.Queryable,
	shop int,
) ([]Report, error) {
	return util.WithCollect[Report](
		ctx,
		base,
		query.ListAllReportsByShopQuery,
		shop,
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

func FindScope(
	ctx context.Context,
	base db.Queryable,
	id any,
) (string, error) {
	var scope string
	err := base.QueryRow(ctx, query.FindReportScopeQuery, id).Scan(&scope)
	return scope, err
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteReportQuery, id)
	return err
}

