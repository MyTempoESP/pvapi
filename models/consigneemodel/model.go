package consigneemodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type Consignee struct {
	ID    int    `db:"consignee_id"`
	Phone string `db:"phone"`

	// anonymous fields
	address.Address
	usermodel.User
}

func (c ConsigneeBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	userID, err := c.User.MustSave(ctx, base)

	if err != nil {
		return
	}

	_, err = model.Create(
		ctx,
		base,
		query.CreateConsigneeQuery,
		util.Merge(
			pgx.NamedArgs{
				"telephone": c.Phone,
				"shop_id":   c.ShopID,
				"user_id":   userID,
			},
			c.Address.Prepare().Params(),
		),
	)

	return
}

func List(
	ctx context.Context,
	pool *pgxpool.Pool,
	shop int,
) ([]Consignee, error) {
	return util.WithCollect[Consignee](
		ctx,
		pool,
		query.ListShopConsigneesQuery,
		shop,
	)
}
