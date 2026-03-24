package usermodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type User struct {
	Name  string `json:"name" db:"user_name"`
	Email string `json:"email" db:"user_email"`
}

func FindOrCreateTx(
	ctx context.Context,
	tx pgx.Tx,
	name, email, password string,
) (userID int, err error) {

	passwordHash, err := util.HashPassword(password)

	if err != nil {
		err = fmt.Errorf("can't create new user's password: %w", err)
		return
	}

	err = tx.QueryRow(
		ctx,
		query.UserCreateIfNotExistsQuery,
		pgx.NamedArgs{
			"name":     name,
			"email":    email,
			"password": passwordHash,
		},
	).Scan(&userID)

	if err == nil {
		return // all ok
	}

	// if user exists (nothing returned on RETURNING id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(
			ctx,
			query.UserFindByEmailQuery,
			email,
		).Scan(&userID)
	}

	return
}

func FindOrCreate(
	ctx context.Context,
	pool *pgxpool.Pool,
	name, email, password string,
) (userID int, err error) {

	err = pgx.BeginFunc(
		ctx,
		pool,
		func(tx pgx.Tx) (err error) {
			// outer, inner =
			userID, err = FindOrCreateTx(ctx, tx, name, email, password)
			return
		},
	)

	return
}

func FindByEmail(
	ctx context.Context,
	pool *pgxpool.Pool,
	email string,
) (userID int, err error) {

	err = pool.QueryRow(
		ctx, query.UserFindByEmailQuery, email,
	).Scan(&userID)

	return
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
