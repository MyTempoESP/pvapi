package productmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
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

	Image      string   `json:"main_image"`
	SpecImages []string `json:"spec_images"`
	DescImages []string `json:"desc_images"`
	PackImages []string `json:"pack_images"`
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
}

func (p ProductBlueprint) Save(
	ctx context.Context,
	base pgx.Tx,
) (err error) {
	_, err = base.Exec(
		ctx,
		query.ProductCreate,
		pgx.NamedArgs{
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
		},
	)
	return
}

func PublicFind(
	ctx context.Context,
	base db.Queryable,
	productID int,
) (Product, error) {
	return util.WithCollectExactlyOneRow[Product](
		ctx,
		base,
		query.ProductPublicFindQuery,
		productID,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	product int,
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
