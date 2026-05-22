package adminmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util/userutils"
)

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,

	adminUser,
	adminEmail,
	adminPass string,
) (err error) {
	user, err := userutils.NewUser(
		adminUser,
		adminEmail,
		adminPass,
	)

	if err != nil {
		return
	}

	err = pgx.BeginFunc(ctx, pool,
		func(tx pgx.Tx) error {
			userID, err := user.MustSave(ctx, tx)

			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, query.AdminCreateQuery, userID)
			return err
		})

	return
}
