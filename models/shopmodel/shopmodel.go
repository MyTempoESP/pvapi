package shopmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"

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

	update.Set.Add("name", "name", shop.Name)
	update.Set.Add("telephone", "telephone", shop.Telephone)

	update.Where.Add("shop.id", "shop_id", shopID)

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

	params := pgx.NamedArgs{
		"user":      user,
		"name":      shop.Name,
		"slug":      slug.Make(shop.Name),
		"telephone": shop.Telephone,

		"addrstreet":     shop.Address.Street,
		"addrnumber":     shop.Address.Number,
		"addrneigh":      shop.Address.Neighborhood,
		"addrcity":       shop.Address.City,
		"addrstate":      shop.Address.State,
		"addrcep":        shop.Address.CEP,
		"addrcomplement": shop.Address.Complement,
	}

	err = pgx.BeginFunc(
		ctx, pool,
		func(tx pgx.Tx) error {
			err := tx.QueryRow(
				ctx,
				query.ShopCreateQuery,
				params,
			).Scan(&shopID)

			return err
		},
	)

	return
}

func Get(
	ctx context.Context,
	pool *pgxpool.Pool,
) (shops []Shop, err error) {
	rows, err := pool.Query(
		ctx, query.SelectAllShopsQuery,
	)

	if err != nil {
		return
	}
	defer rows.Close()

	shops, err = pgx.CollectRows(
		rows, pgx.RowToStructByName[Shop],
	)

	if err != nil {
		return
	}

	err = rows.Err()

	return
}

func Find(
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
) (shop Shop, err error) {
	row, err := pool.Query(
		ctx, query.ShopFindQuery,
		id,
	)

	if err != nil {
		return
	}
	defer row.Close()

	shop, err = pgx.CollectExactlyOneRow(
		row, pgx.RowToStructByName[Shop],
	)

	if err != nil {
		return
	}

	err = row.Err()

	return
}

// FindDetailed looks for the products count and
// pickup locations count as extra fields
func FindDetailed(
	ctx context.Context,
	pool *pgxpool.Pool,
	id string,
) (shop DetailedShop, err error) {
	row, err := pool.Query(
		ctx, query.ShopFindDetailedQuery,
		id,
	)

	if err != nil {
		return
	}
	defer row.Close()

	shop, err = pgx.CollectExactlyOneRow(
		row, pgx.RowToStructByName[DetailedShop],
	)

	if err != nil {
		return
	}

	err = row.Err()

	return
}
