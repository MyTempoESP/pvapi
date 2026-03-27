package pickupmodel

import (
	"context"
	"time"

	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type PickupLocation struct {
	ID       int
	Name     string
	Datetime time.Time `db:"pickup_at"`
	Address  *address.Address
}

func List(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]PickupLocation, error) {
	return util.WithCollect[PickupLocation](
		ctx,
		base,
		query.ListPickupLocationsQuery,
		shopID,
	)
}

func (p PickupLocationBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	return model.Create(
		ctx,
		base,
		query.CreatePickupLocationQuery,
		util.Merge(
			util.Args{
				"name":      p.Name,
				"shop_id":   p.ShopID,
				"pickup_at": p.Datetime,
			},
			p.Address.Params(),
		),
	)
}
