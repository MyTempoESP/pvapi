package shopmodel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/categorymodel"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"

	"github.com/gosimple/slug"
)

type Shop struct {
	ID        int    `json:"id" db:"shop_id"`
	Name      string `json:"name" db:"name"`
	Slug      string `json:"slug" db:"slug"`
	Telephone string `json:"phone" db:"telephone"`
	Status    string `json:"status" db:"status"`

	address.Address
	usermodel.User
}

type DetailedShop struct {
	ID        int    `json:"id" db:"shop_id"`
	Name      string `json:"name" db:"name"`
	Slug      string `json:"slug" db:"slug"`
	Telephone string `json:"phone" db:"telephone"`
	Status    string `json:"status" db:"status"`
	Products  int    `json:"products_count" db:"products_count"`
	Pickups   int    `json:"pickup_locations_count" db:"pickups"`

	address.Address
	usermodel.User
}

func FindByUser(
	ctx context.Context,
	base db.Queryable,
	userID int,
) (Shop, error) {
	return util.WithCollectExactlyOneRow[Shop](
		ctx,
		base,
		query.ShopFindByUserQuery,
		userID,
	)
}

func FindByUserDetailed(
	ctx context.Context,
	base db.Queryable,
	userID int,
) (DetailedShop, error) {
	return util.WithCollectExactlyOneRow[DetailedShop](
		ctx,
		base,
		query.ShopFindDetailedByUserQuery,
		userID,
	)
}

func Update(
	ctx context.Context,
	pool *pgxpool.Pool,
	shopID int,
	shop UpdateShopRequest,
) (err error) {
	update := query.Update("shops AS shop")

	if shop.Email != nil {
		err = pgx.BeginFunc(
			ctx, pool, func(tx pgx.Tx) error {
				_, err := tx.Exec(
					ctx,
					query.ShopUserUpdateEmail,
					shopID,
					shop.Email,
				)
				return err
			},
		)

		if err != nil {
			return
		}
	}

	update.Set.Add("name", shop.Name)
	update.Set.Add("telephone", shop.Telephone)

	update.Where.AddMap("shop.id", "shop_id", shopID)

	query, params := update.Build()

	err = pgx.BeginFunc(
		ctx, pool,
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, query, params)
			return err
		},
	)

	return
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	shop CreateShopRequest,
) (shopID int, err error) {
	user, err := usermodel.FindOrCreate(
		ctx, pool,
		shop.Name,
		shop.Email,
		shop.Telephone, // telephone as password if does not exist
	)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	var addrID int

	// TODO: make this a func
	addrParams := pgx.NamedArgs{
		"street":     shop.Address.Street,
		"number":     shop.Address.Number,
		"neigh":      shop.Address.Neighborhood,
		"city":       shop.Address.City,
		"state":      shop.Address.State,
		"cep":        shop.Address.CEP,
		"complement": shop.Address.Complement,
	}

	err = tx.QueryRow(
		ctx,
		query.CreateAddressQuery,
		addrParams,
	).Scan(&addrID)

	if err != nil {
		return
	}

	shopParams := pgx.NamedArgs{
		"user":      user,
		"name":      shop.Name,
		"slug":      slug.Make(shop.Name),
		"address":   addrID,
		"telephone": shop.Telephone,
	}

	err = tx.QueryRow(
		ctx,
		query.ShopCreateQuery,
		shopParams,
	).Scan(&shopID)

	if err != nil {
		return
	}

	err = categorymodel.CreateTx(
		ctx, tx,
		"n/a",
		"Sem categoria",
		shopID,
	)

	if err != nil {
		err = fmt.Errorf(
			"couldn't create default category: %w",
			err,
		)
	}

	err = tx.Commit(ctx)

	return
}

func List(
	ctx context.Context,
	pool *pgxpool.Pool,
) ([]Shop, error) {
	return util.WithCollect[Shop](
		ctx,
		pool,
		query.SelectAllShopsQuery,
	)
}

func Find(
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
) (Shop, error) {
	return util.WithCollectExactlyOneRow[Shop](
		ctx,
		pool,
		query.ShopFindQuery,
		id,
	)
}

// FindDetailed looks for the products count and
// pickup locations count as extra fields
func FindDetailed(
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
) (DetailedShop, error) {
	return util.WithCollectExactlyOneRow[DetailedShop](
		ctx,
		pool,
		query.ShopFindDetailedQuery,
		id,
	)
}
