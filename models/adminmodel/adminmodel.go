package adminmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
)

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,

	adminUser,
	adminEmail,
	adminPass string,
) (err error) {
	userID, err := usermodel.Create(
		ctx, pool,
		adminUser,
		adminEmail,
		adminPass,
	)

	if err != nil {
		return
	}

	err = pgx.BeginFunc(ctx, pool,
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, query.AdminCreateQuery, userID)
			return err
		})

	return
}
