package productmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/models/imagemodel"
	"linhazero.com/pvapi/query"
)

const (
	defaultImageTag       = "default"
	descriptionImageTag   = "description"
	specificationImageTag = "specification"
	deliveryImageTag      = "delivery"
)

type Specification struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Policy struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type EarningsModifier struct {
	Type  string          `json:"type"`
	Value decimal.Decimal `json:"value"`
}

type ProductAttr struct {
	Colors []string `json:"colors"`
	Sizes  []string `json:"sizes"`
}

type Product struct {
	ID               int             `db:"product_id"`
	Name             string          `db:"product_name"`
	Price            decimal.Decimal `db:"product_price"`
	Stock            int             `db:"product_stock"`
	Discount         decimal.Decimal `db:"product_discount"`
	Featured         bool            `db:"product_featured"`
	ModerationStatus string          `db:"product_moderation_status"`
	Description      string          `db:"product_description"`
	ShopID           int             `db:"product_shop_id"`
	Category         string          `db:"product_category"`
	UUID             string          `db:"product_uuid"`
	Code             string          `db:"product_code"`

	Attributes ProductAttr `db:"product_attr"`

	Specifications []Specification `db:"product_specifications"`

	AllowCustomName           bool `db:"product_p_cname"`
	AllowCustomModality       bool `db:"product_p_cmod"`
	AllowCustomColorSelection bool `db:"product_p_ccolor"`

	// nullable
	PlatformCut *EarningsModifier `db:"product_platform_cut"`

	Image      string   `db:"product_main_image"`
	SpecImages []string `db:"product_spec_images"`
	DescImages []string `db:"product_desc_images"`
	PackImages []string `db:"product_pack_images"`
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
		"name":        product.Name,
		"price":       product.Price,
		"stock":       product.Stock,
		"discount":    product.Discount,
		"featured":    product.Featured,
		"description": product.Description,
		"code":        product.Code,
		"category":    product.Category,
		"shop_id":     shop,
		"attr":        product.Attributes,
	}

	err = tx.QueryRow(
		ctx,
		query.ProductCreate,
		params,
	).Scan(&productID)

	if err != nil {
		return
	}

	err = imagemodel.NewImage(
		imagemodel.URL(product.ImageURL),
		defaultImageTag, productID,
	).Create(ctx, tx)

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
			pol.Enabled,
			productID,
		)
	}

	imagemodel.NewTaggedImageCollection(
		product.DescImages,
		descriptionImageTag,
		productID,
	).Queue(batch)

	imagemodel.NewTaggedImageCollection(
		product.SpecImages,
		specificationImageTag,
		productID,
	).Queue(batch)

	imagemodel.NewTaggedImageCollection(
		product.PackImages,
		deliveryImageTag,
		productID,
	).Queue(batch)

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
