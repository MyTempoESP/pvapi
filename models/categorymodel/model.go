package categorymodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/util"
)

type Category struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	ShopID      int    `db:"shop_id"`
	Description string `db:"description"`
}

func ListByShop(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]Category, error) {
	return util.WithCollect[Category](
		ctx, base,
		query.CategoryListByShopQuery,
		shopID,
	)
}

func CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	name, desc string,
	shop int,
) (id int, err error) {
	err = tx.QueryRow(
		ctx, query.CreateCategoryQuery,
		name, desc, shop,
	).Scan(&id)

	return
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	name, desc string,
	shop int,
) (err error) {
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		_, err := CreateTx(ctx, tx, name, desc, shop)
		return err
	})

	return
}

