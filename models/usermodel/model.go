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
	Highest      string   `json:"highest"`
	ActiveRoleID int      `json:"active_role_id"`
	AllRoles     []string `json:"all_roles"`
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

	var roles []string

	if levels.Admin != nil {
		roles = append(roles, "admin")
	}
	if levels.Shop != nil {
		roles = append(roles, "shop")
	}
	if levels.Consignee != nil {
		roles = append(roles, "consignee")
	}
	if levels.Client != nil {
		roles = append(roles, "client")
	}
	if len(roles) == 0 {
		roles = append(roles, "user")
	}

	var p Permissions
	p.AllRoles = roles

	if levels.Admin != nil {
		p.Highest = "admin"
		p.ActiveRoleID = *levels.Admin
	} else if levels.Shop != nil {
		p.Highest = "shop"
		p.ActiveRoleID = *levels.Shop
	} else if levels.Consignee != nil {
		p.Highest = "consignee"
		p.ActiveRoleID = *levels.Consignee
	} else if levels.Client != nil {
		p.Highest = "client"
		p.ActiveRoleID = *levels.Client
	} else {
		p.Highest = "user"
		p.ActiveRoleID = user
	}

	return p, nil
}
