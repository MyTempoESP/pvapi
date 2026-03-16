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

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	user int,
	shop CreateShopRequest,
) (shopID int, err error) {

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

func (shop Shop) ToResource() ShopResource {
	return ShopResource{
		ID:        shop.ID,
		Name:      shop.Name,
		Slug:      shop.Slug,
		Telephone: shop.Telephone,
		Status:    shop.Status,

		Address: shop.Address,
		User:    shop.User,
	}
}
