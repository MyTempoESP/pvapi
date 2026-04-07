package consignmentmodel

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"

	//"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

func (a AddProductAction) Exec(
	ctx context.Context,
	tx pgx.Tx,
) (err error) {
	batch := &pgx.Batch{}

	for _, p := range a.Products {
		product := p.Prepare()

		batch.Queue(
			query.ProductAddToConsignmentQuery,
			util.Args{
				"product_id":     product.ID,
				"consignment_id": a.ConsignmentID,
				"quantity":       product.Quantity,
				"price":          product.Price,
				"attr":           product.AttrVal,
				"discount":       product.Discount,
				"featured":       product.Featured,
			},
		)
	}

	err = util.Bake(ctx, tx, batch)

	if err != nil {
		return fmt.Errorf(
			"Couldn't add products to consignment %q: %w",
			a.ConsignmentID,
			err,
		)
	}

	return
}

func (c ConsignmentUpdateBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	q := util.SQL

	fields := make(map[string]any)

	if c.PickupID != nil {
		fields["pickup_location_id"] = *c.PickupID
	}

	if c.ConsigneeID != nil {
		fields["consignee_id"] = *c.ConsigneeID
	}

	if c.ComissionType != nil {
		fields["comission_type"] = *c.ComissionType
	}

	if c.ComissionValue != nil {
		fields["comission_value"] = *c.ComissionValue
	}

	if c.Status != nil {
		fields["status"] = *c.Status
	}

	if len(fields) == 0 {
		return
	}

	query, params, err := q.Update("consignments").
		SetMap(fields).
		Where(sq.Eq{"id": c.ID}).
		ToSql()

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

func (c ConsignmentBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	return model.Create(
		ctx,
		base,
		query.CreateConsignmentQuery,
		util.Args{
			"shop_id":            c.ShopID,
			"pickup_location_id": c.PickupID,
			"consignee_id":       c.ConsigneeID,
			"comission_type":     c.ComissionType,
			"comission_value":    c.ComissionValue,
			"status":             c.Status,
		},
	)
}

func List(
	ctx context.Context,
	base db.Queryable,
) ([]Consignment, error) {
	return util.WithCollect[Consignment](
		ctx,
		base,
		query.ListAllConsignmentsQuery,
	)
}

func ListProducts(
	ctx context.Context,
	base db.Queryable,
	id any,
) ([]Product, error) {
	return util.WithCollect[Product](
		ctx,
		base,
		query.ListAllConsignmentProductsQuery,
		id,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (DetailedConsignment, error) {
	return util.WithCollectExactlyOneRow[DetailedConsignment](
		ctx,
		base,
		query.FindConsignmentQuery,
		id,
	)
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteConsignmentQuery, id)
	return err
}
