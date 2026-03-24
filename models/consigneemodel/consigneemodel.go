package consigneemodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/address"
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

func CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	consignee *ConsigneeBlueprint,
) (err error) {

	userID, err := usermodel.FindOrCreateTx(
		ctx,
		tx,
		consignee.Name,
		consignee.Email,
		consignee.Phone, // phone as password by default
	)

	_, err = tx.Exec(
		ctx,
		query.CreateConsigneeQuery,
		pgx.NamedArgs{
			"telephone": consignee.Phone,
			"shop_id":   consignee.ShopID,
			"user_id":   userID,

			"addrstreet":     consignee.Address.Street,
			"addrnumber":     consignee.Address.Number,
			"addrneigh":      consignee.Address.Neighborhood,
			"addrcity":       consignee.Address.City,
			"addrstate":      consignee.Address.State,
			"addrcep":        consignee.Address.CEP,
			"addrcomplement": consignee.Address.Complement,
		},
	)

	return
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	consignee *ConsigneeBlueprint,
) (err error) {

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		return CreateTx(ctx, tx, consignee)
	})

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
