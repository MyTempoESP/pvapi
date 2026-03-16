package productmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type Product struct {
	ID       int             `db:"product_id"`
	Name     string          `db:"product_name"`
	Code     string          `db:"product_code"`
	Price    decimal.Decimal `db:"product_price"`
	Stock    int             `db:"product_stock"`
	Discount decimal.Decimal `db:"product_discount"`

	ProductSpecifications
}

func GetByShop(
	ctx context.Context,
	pool *pgxpool.Pool,
	shopID string,
) (products []Product, err error) {

	rows, err := pool.Query(
		ctx,
		query.ProductGetByShop,
		shopID,
	)

	if err != nil {
		return
	}
	defer rows.Close()

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[Product])

	return
}
