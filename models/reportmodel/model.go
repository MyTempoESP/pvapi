package reportmodel

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
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
	OrderName     string    `db:"order_name"`
	CreatedAt     time.Time `db:"created_at"`
}

func (b UpdateReportBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) error {
	q := util.SQL

	fields := make(map[string]any)

	if b.Status != nil {
		fields["status"] = *b.Status
	}
	if b.Priority != nil {
		fields["priority"] = *b.Priority
	}
	if b.Description != nil {
		fields["description"] = *b.Description
	}
	if b.Type != nil {
		fields["type"] = *b.Type
	}

	if len(fields) == 0 {
		return nil
	}

	update := q.Update("reports").
		SetMap(fields).
		Where(sq.Eq{"id": b.ID})

	query, params, err := update.ToSql()
	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	_, err = base.Exec(ctx, query, params...)
	if err != nil {
		return fmt.Errorf("could not update report, query: %q, %w", query, err)
	}

	return nil
}

func createBaseReportFields(
	description,
	moment,
	scope,
	priority,
	status,
	r_type string,
) util.Args {
	return util.Args{
		"description": description,
		"moment":      moment,
		"scope":       scope,
		"priority":    priority,
		"status":      status,
		"type":        r_type,
	}
}

func newReport(
	ctx context.Context,
	base db.Queryable,
	fields util.Args,
	scope string,
	orderID int,
) (int, error) {
	if scope == "product" {
		// assign product_id if scope is product
		fields["product_id"] = orderID

		return model.Create(
			ctx,
			base,
			query.CreateProductReportQuery,
			fields,
		)
	}

	// assign consignment_id if scope is consignment
	fields["consignment_id"] = orderID

	return model.Create(
		ctx,
		base,
		query.CreateReportQuery,
		fields,
	)
}

func (r ReportWithConsigneeIDBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	fields := createBaseReportFields(
		r.Description,
		r.Moment,
		r.Scope,
		r.Priority,
		r.Status,
		r.Type,
	)

	fields["consignee_id"] = r.ConsigneeID

	if r.Scope == "product" {
		fields["product_id"] = int(r.OrderID)
		return model.Create(ctx, base, query.CreateProductReportWithConsigneeIDQuery, fields)
	}

	fields["consignment_id"] = int(r.OrderID)
	return model.Create(ctx, base, query.CreateReportWithConsigneeIDQuery, fields)
}

func (r ReportBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	fields := createBaseReportFields(
		r.Description,
		r.Moment,
		r.Scope,
		r.Priority,
		r.Status,
		r.Type,
	)

	// assign consignee id in this case
	fields["consignee_name"] = r.Monitor

	return newReport(ctx, base, fields, r.Scope, int(r.OrderID))
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
