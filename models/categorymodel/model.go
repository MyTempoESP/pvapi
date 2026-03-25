package categorymodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/query"
)

func CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	name, desc string,
	shop int,
) (err error) {
	_, err = tx.Exec(
		ctx, query.CreateCategoryQuery,
		name, desc, shop,
	)

	return
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	name, desc string,
	shop int,
) (err error) {
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		err := CreateTx(ctx, tx, name, desc, shop)
		return err
	})

	return
}
