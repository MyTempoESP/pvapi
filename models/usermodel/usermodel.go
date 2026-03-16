package usermodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type User struct {
	Name  string `json:"name" db:"user_name"`
	Email string `json:"email" db:"user_email"`
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,

	userName,
	userEmail,
	userPassword string,
) (userID int, err error) {
	passwordHash, err := util.HashPassword(userPassword)

	// TODO: Check error type, password too long etc.
	if err != nil {
		return
	}

	// rollback!
	err = pgx.BeginFunc(
		ctx, pool,
		func(tx pgx.Tx) error {
			err := tx.QueryRow(
				ctx,
				query.UserRegisterQuery,
				userName, userEmail, passwordHash,
			).Scan(&userID)

			return err
		},
	)

	return
}

func Login(
	ctx context.Context,
	pool *pgxpool.Pool,
	userMail string,
	password string,
) (userID int, err error) {

	var hashed string

	err = pool.QueryRow(
		ctx,
		query.UserPasswordQuery, userMail,
	).Scan(&userID, &hashed)

	if err != nil {
		return
	}

	err = util.ComparePasswordWithHash(password, []byte(hashed))

	return
}

func GetUserPermissionLevel(
	ctx context.Context,
	pool *pgxpool.Pool,
	user int,
) (level string, err error) {
	var (
		admin,
		shop,
		consignee bool

		_user int
	)

	err = pool.QueryRow(
		ctx,
		query.UserMaxLevelQuery, user,
	).Scan(&admin, &shop, &consignee, &_user)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}

		return
	}

	if admin {
		level = "admin"
		return
	}

	if shop {
		level = "shop"
		return
	}

	if consignee {
		level = "consignee"
		return
	}

	level = "none"

	return
}
