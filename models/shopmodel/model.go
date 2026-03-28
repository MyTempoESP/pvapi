package shopmodel

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/userutils"

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
	q := util.SQL

	fields := make(map[string]any)

	if shop.Name != nil {
		fields["name"] = *shop.Name
	}

	if shop.Telephone != nil {
		fields["telephone"] = *shop.Telephone
	}

	if len(fields) == 0 {
		return
	}

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

	query, params, err := q.Update("shops").
		SetMap(fields).
		Where(sq.Eq{"id": shopID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	err = pgx.BeginFunc(
		ctx, pool,
		func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, query, params...)
			return err
		},
	)

	return
}

func CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	shop CreateShopRequest,
	user userutils.User,
) (shopID int, err error) {

	userID, err := user.MustSave(ctx, tx)

	if err != nil {
		return
	}

	shopParams := util.Merge(
		pgx.NamedArgs{
			"user":      userID,
			"name":      shop.Name,
			"slug":      slug.Make(shop.Name),
			"telephone": shop.Telephone,
		},
		shop.Address.Prepare().Params(),
	)

	err = tx.QueryRow(
		ctx,
		query.ShopCreateQuery,
		shopParams,
	).Scan(&shopID)

	return
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	shop CreateShopRequest,
) (shopID int, err error) {
	user, err := userutils.NewUser(
		shop.Name,
		shop.Email,
		shop.Telephone,
	)

	if err != nil {
		return
	}

	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) (err error) {
		shopID, err = CreateTx(ctx, tx, shop, user)
		return
	})

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
	base db.Queryable,
	id int,
) (Shop, error) {
	return util.WithCollectExactlyOneRow[Shop](
		ctx,
		base,
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

func HasResource(
	ctx context.Context,
	base db.Queryable,
	query string,
	shop, id any,
) (hasResource bool, err error) {
	err = base.QueryRow(
		ctx,
		query,
		shop, id,
	).Scan(&hasResource)

	return
}

func HasProduct(
	ctx context.Context,
	base db.Queryable,
	shop, product any,
) (bool, error) {
	return HasResource(
		ctx,
		base,
		query.ShopHasProductQuery,
		shop,
		product,
	)
}

func HasPickup(
	ctx context.Context,
	base db.Queryable,
	shop, pickup any,
) (bool, error) {
	return HasResource(
		ctx,
		base,
		query.ShopHasPickupQuery,
		shop,
		pickup,
	)
}

func HasConsignment(
	ctx context.Context,
	base db.Queryable,
	shop, consignment any,
) (bool, error) {
	return HasResource(
		ctx,
		base,
		query.ShopHasConsignmentQuery,
		shop,
		consignment,
	)
}

func HasConsignee(
	ctx context.Context,
	base db.Queryable,
	shop, consignee any,
) (bool, error) {
	return HasResource(
		ctx,
		base,
		query.ShopHasConsigneeQuery,
		shop,
		consignee,
	)
}
