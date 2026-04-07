package productmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/imagemodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
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

type ProductAttr struct {
	Colors     []string `json:"colors,omitempty"`
	Sizes      []string `json:"sizes,omitempty"`
	Modalities []string `json:"modalities,omitempty"`

	AllowCustomColor    bool `json:"customColor,omitempty"`
	AllowCustomModality bool `json:"customModality,omitempty"`
	AllowCustomName     bool `json:"customName,omitempty"`

	Specifications []Specification `json:"specifications"`
}

type Product struct {
	ShopID           int             `db:"product_shop_id"`
	ShopName         string          `db:"product_shop_name"`
	ID               int             `db:"product_id"`
	Name             string          `db:"product_name"`
	Price            decimal.Decimal `db:"product_price"`
	Stock            int             `db:"product_stock"`
	Discount         decimal.Decimal `db:"product_discount"`
	Featured         bool            `db:"product_featured"`
	ModerationStatus string          `db:"product_moderation_status"`
	Description      string          `db:"product_description"`
	Category         string          `db:"product_category"`
	UUID             string          `db:"product_uuid"`
	Code             string          `db:"product_code"`
	Attributes       ProductAttr     `db:"product_attr"`
	PlatformCutType  string          `db:"product_platform_cut_type"`
	PlatformCut      decimal.Decimal `db:"product_platform_cut"`

	Image      string   `db:"product_main_image"`
	SpecImages []string `db:"product_spec_images"`
	DescImages []string `db:"product_desc_images"`
	PackImages []string `db:"product_pack_images"`
}

func (p ProductBlueprint) CreateProductInfo(
	ctx context.Context,
	base pgx.Tx,
	productID int,
) (err error) {
	err = imagemodel.NewImage(
		imagemodel.NewUrl(p.ImageURL),
		defaultImageTag, productID,
	).Create(ctx, base)

	if err != nil {
		return
	}

	batch := &pgx.Batch{}

	imagemodel.NewTaggedImageCollection(
		p.DescImages,
		descriptionImageTag,
		productID,
	).Queue(batch)

	imagemodel.NewTaggedImageCollection(
		p.SpecImages,
		specificationImageTag,
		productID,
	).Queue(batch)

	imagemodel.NewTaggedImageCollection(
		p.PackImages,
		deliveryImageTag,
		productID,
	).Queue(batch)

	err = util.Bake(ctx, base, batch)

	return
}

func (p ProductBlueprint) Save(
	ctx context.Context,
	base pgx.Tx,
) (err error) {
	params := pgx.NamedArgs{
		"name":        p.Name,
		"price":       p.Price,
		"stock":       p.Stock,
		"discount":    p.Discount,
		"featured":    p.Featured,
		"description": p.Description,
		"code":        p.Code,
		"category":    p.Category,
		"shop_id":     p.Shop,
		"attr":        p.Attributes,
	}

	var productID int

	err = base.QueryRow(ctx, query.ProductCreate, params).Scan(&productID)

	if err != nil {
		return
	}

	err = p.CreateProductInfo(ctx, base, productID)

	return
}

func Find(
	ctx context.Context,
	base db.Queryable,
	product any, // id
) (p Product, err error) {
	err = base.QueryRow(ctx, query.FindProductQuery, product).Scan(&p)
	return
}

func Take(
	ctx context.Context,
	base db.Queryable,
	product,
	ammount int,
) (err error) {
	_, err = base.Exec(
		ctx,
		query.ProductTakeFromStock,
		product,
		ammount,
	)
	return
}

func PublicIndexByShop(
	ctx context.Context,
	pool *pgxpool.Pool,
	shopID any,
) (products []Product, err error) {

	rows, err := pool.Query(
		ctx,
		query.ProductPublicIndexByShopQuery,
		shopID,
	)

	if err != nil {
		return
	}
	defer rows.Close()

	products, err = pgx.CollectRows(rows, pgx.RowToStructByName[Product])

	return
}

func PendingIndex(
	ctx context.Context,
	base db.Queryable,
) ([]Product, error) {
	return util.WithCollect[Product](
		ctx,
		base,
		query.ProductsPendingModerationQuery,
	)
}

func Moderate(
	ctx context.Context,
	base db.Queryable,
	request ModerateProductRequest,
	productID int,
	newStatus string,
) (err error) {
	_, err = base.Exec(
		ctx,
		query.ProductModerate,
		util.Args{
			"id":                productID,
			"status":            newStatus,
			"platform_cut":      request.PlatformCut,
			"platform_cut_type": request.PlatformCutType,
		},
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

// TODO: Update product (PAIN)
