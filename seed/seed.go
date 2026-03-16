package seed

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/env"
	"linhazero.com/pvapi/models/adminmodel"
)

func Plant(
	logger *slog.Logger,
	ctx context.Context,
	pool *pgxpool.Pool,
) (err error) {
	adminUser, err := env.AdminUser()

	if err != nil {
		return
	}

	adminPass, err := env.AdminPass()

	if err != nil {
		return
	}

	adminMail, err := env.AdminMail()

	if err != nil {
		return
	}

	logger.Debug("Creating admin user",
		"user", adminUser, "pass", adminPass)

	adminmodel.Create(
		ctx, pool,
		adminUser,
		adminMail,
		adminPass,
	)

	return
}
