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
	Highest      string `db:"-" json:"highest"`
	ActiveRoleID int    `db:"-" json:"active_role_id"`

	Admin     *int `db:"admin"     json:"admin,omitempty"`
	Shop      *int `db:"shop"      json:"shop,omitempty"`
	Consignee *int `db:"consignee" json:"monitor,omitempty"`
	Client    *int `db:"client"    json:"cliente,omitempty"`
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

	p, err := util.WithCollectExactlyOneRow[Permissions](
		ctx,
		pool,
		query.UserPermissionsQuery,
		user,
	)

	if err != nil {
		return zero, err
	}

	if p.Admin != nil {
		p.Highest = "admin"
		p.ActiveRoleID = *p.Admin
	} else if p.Shop != nil {
		p.Highest = "shop"
		p.ActiveRoleID = *p.Shop
	} else if p.Consignee != nil {
		p.Highest = "consignee"
		p.ActiveRoleID = *p.Consignee
	} else if p.Client != nil {
		p.Highest = "client"
		p.ActiveRoleID = *p.Client
	} else {
		p.Highest = "user"
		p.ActiveRoleID = user
	}

	return p, nil
}
