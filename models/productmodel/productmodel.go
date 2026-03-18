package productmodel

import (
	"context"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/query"
)

const (
	defaultImageTag       = "default"
	descriptionImageTag   = "description"
	specificationImageTag = "specification"
	deliveryImageTag      = "delivery"
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
	Price    decimal.Decimal `db:"product_Price"`
	Stock    int             `db:"product_stock"`
	Discount decimal.Decimal `db:"product_discount"`
	Featured bool            `db:"product_featured"`

	Specifications []Specification `db:"product_specifications"`
}

func createImages(
	tag string,
	images []URL,
	batch *pgx.Batch,
	product int,
) {
	for _, url := range images {
		batch.Queue(
			query.CreateImageQuery,
			url,
			slug.Make(string(url)),
			tag,
			product,
		)
	}
}

func productTransaction(
	ctx context.Context,
	pool *pgxpool.Pool,
	product *ProductBlueprint,
	shop int,
) (err error) {
	var productID int

	tx, err := pool.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	params := pgx.NamedArgs{
		"name":     product.Name,
		"price":    product.Price,
		"stock":    product.Stock,
		"discount": product.Discount,
		"featured": product.Featured,
		"shop_id":  shop,
	}

	err = tx.QueryRow(
		ctx,
		query.ProductCreate,
		params,
	).Scan(&productID)

	if err != nil {
		return
	}

	batch := &pgx.Batch{}

	// specifications
	for _, spec := range product.Specifications {
		batch.Queue(
			query.CreateSpecificationQuery,
			spec.Name,
			spec.Value,
			productID,
		)
	}

	for _, pol := range product.Policies {
		batch.Queue(
			query.CreatePolicyQuery,
			pol.Name,
			pol.Value,
			pol.Enabled,
			productID,
		)
	}

	batch.Queue(
		query.CreateImageQuery,
		product.ImageURL,
		slug.Make(string(product.ImageURL)),
		defaultImageTag,
		productID,
	)

	createImages(deliveryImageTag, product.PackImages, batch, productID)
	createImages(specificationImageTag, product.SpecImages, batch, productID)
	createImages(descriptionImageTag, product.DescImages, batch, productID)

	b := tx.SendBatch(ctx, batch)
	defer b.Close()

	for range batch.Len() {
		_, err = b.Exec()

		if err != nil {
			return
		}
	}

	err = b.Close()

	if err != nil {
		return
	}

	err = tx.Commit(ctx)

	return
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	product *ProductBlueprint,
	shop int,
) (productID int, err error) {

	err = productTransaction(
		ctx, pool,
		product, shop,
	)

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
