package consignmentmodel

import (
	"context"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

func (c ConsignmentUpdateBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	update := query.Update("consignments")

	update.Set.Add("pickup_location_id", c.PickupID)
	update.Set.Add("consignee_id", c.ConsigneeID)
	update.Set.Add("comission_type", c.ComissionType)
	update.Set.Add("comission_value", c.ComissionValue)
	update.Set.Add("status", c.Status)

	update.Where.AddMap("id", "consignment_id", c.ID)

	query, params := update.Build()

	_, err = base.Exec(ctx, query, params)

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
