package usermodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type User struct {
	Name  string `json:"name" db:"user_name"`
	Email string `json:"email" db:"user_email"`
}

// FindOrCreate is a utility function, not a controller resource
// and as such, it should be used with care (e.g. remembering to initiate a tx)
func FindOrCreate(
	ctx context.Context,
	base db.Queryable,
	name, email string,
	passwordHash []byte,
) (userID int, err error) {

	err = base.QueryRow(
		ctx,
		query.UserFindOrCreateQuery,
		pgx.NamedArgs{
			"name":     name,
			"email":    email,
			"password": passwordHash,
		},
	).Scan(&userID)

	return
}

func FindByEmail(
	ctx context.Context,
	base db.Queryable,
	email string,
) (userID int, err error) {

	err = base.QueryRow(
		ctx, query.UserFindByEmailQuery, email,
	).Scan(&userID)

	return
}

func Create(
	ctx context.Context,
	base db.Queryable,
	name, email string,
	passwordHash []byte,
) (userID int, err error) {
	err = base.QueryRow(
		ctx,
		query.UserRegisterQuery,
		name, email, passwordHash,
	).Scan(&userID)

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
