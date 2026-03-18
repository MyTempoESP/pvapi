package productmodel

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/query"
)

type Image struct {
	Filename string `json:"image_name"`
	Alt      string `json:"image_desc"`
}

type Specification struct {
	Name   string  `json:"spec_name"`
	Value  string  `json:"spec_value"`
	Images []Image `json:"spec_images"`
}

type Product struct {
	ID       int             `db:"product_id"`
	Name     string          `db:"product_name"`
	Price    decimal.Decimal `db:"product_price"`
	Stock    int             `db:"product_stock"`
	Discount decimal.Decimal `db:"product_discount"`
	Featured bool            `db:"product_featured"`

	Specifications []Specification `db:"product_specifications"`
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	product CreateProductRequest,
) (productID int, err error) {
	uuid, err := uuid.NewV4()

	if err != nil {
		return
	}

	return
}

func GetByShop(
	ctx context.Context,
	pool *pgxpool.Pool,
	shopID any,
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
