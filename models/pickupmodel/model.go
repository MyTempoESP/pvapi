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

	address.Address
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

func Find(
	ctx context.Context,
	base db.Queryable,
	id int,
) (PickupLocation, error) {
	return util.WithCollectExactlyOneRow[PickupLocation](
		ctx, base,
		query.FindPickupLocationQuery,
		id,
	)
}

func Update(
	ctx context.Context,
	base db.Queryable,
	pickupID int,
	req UpdatePickupLocationRequest,
) error {
	var name = req.Name

	var pickupAt any
	if req.DateTime != nil {
		t, err := time.Parse("2006-01-02T15:04:05", *req.DateTime)
		if err != nil {
			return err
		}
		pickupAt = t
	}

	_, err := base.Exec(
		ctx,
		query.UpdatePickupLocationQuery,
		pickupID,
		name,
		pickupAt,
	)
	if err != nil {
		return err
	}

	if req.Address != nil {
		a, err := req.Address.Prepare()
		if err != nil {
			return err
		}

		_, err = base.Exec(
			ctx,
			query.UpdateAddressQuery,
			pickupID,
			a.Street,
			a.Number,
			a.Neigh,
			a.City,
			a.State,
			a.CEP,
			a.Complement,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id int,
) error {
	_, err := base.Exec(ctx, query.DeletePickupLocationQuery, id)
	return err
}
