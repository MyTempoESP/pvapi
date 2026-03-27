package consignmentmodel

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type Consignee struct {
	ID   int    `db:"consignee_id"`
	Name string `db:"consignee_name"`
}

type Shop struct {
	ID   int    `db:"shop_id"`
	Name string `db:"shop_name"`
}

type Pickup struct {
	ID   int    `db:"pickup_location_id"`
	Name string `db:"pickup_location_name"`
}

type Consignment struct {
	ID             int             `db:"consignment_id"`
	ComissionType  string          `db:"consignment_comission_type"`
	ComissionValue decimal.Decimal `db:"consignment_comission_value"`
	Status         string          `db:"consignment_status"`
	CreatedAt      time.Time       `db:"consignment_created_at"`
	Slug           string          `db:"consignment_slug"`
	Code           string          `db:"consignment_code"`

	Shop
	Consignee
	Pickup
}

func (c Consignment) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	// TODO
	/*
		update := query.Update("consignments AS consignment")

		update.Set.Add("field", c.Field)

		update.Where.AddMap("consignment.id", "consignment_id", c.ID)

		query, params := update.Build()

		_, err = base.Exec(ctx, query, params)

		return err
	*/
	return util.NotImplemented("Consignee Update")
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

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (Consignment, error) {
	return util.WithCollectExactlyOneRow[Consignment](
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
