package usermodel

import (
	"context"

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

type Permissions struct {
	Highest      string
	ActiveRoleID int
}

type permissionOutput struct {
	Admin     *int `db:"admin"`
	Shop      *int `db:"shop"`
	Consignee *int `db:"consignee"`
	Client    *int `db:"client"`
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

func GetUserPermissions(
	ctx context.Context,
	pool *pgxpool.Pool,
	user int,
) (Permissions, error) {
	var zero Permissions

	levels, err := util.WithCollectExactlyOneRow[permissionOutput](
		ctx,
		pool,
		query.UserPermissionsQuery,
		user,
	)

	if err != nil {
		return zero, err
	}

	if levels.Admin != nil {
		return Permissions{
			Highest:      "admin",
			ActiveRoleID: *levels.Admin,
		}, nil
	}

	if levels.Shop != nil {
		return Permissions{
			Highest:      "shop",
			ActiveRoleID: *levels.Shop,
		}, nil
	}

	if levels.Consignee != nil {
		return Permissions{
			Highest:      "consignee",
			ActiveRoleID: *levels.Consignee,
		}, nil
	}

	if levels.Client != nil {
		return Permissions{
			Highest:      "client",
			ActiveRoleID: *levels.Client,
		}, nil
	}

	return Permissions{
		Highest:      "user",
		ActiveRoleID: user,
	}, nil
}
