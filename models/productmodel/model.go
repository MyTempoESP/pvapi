package productmodel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"

	sq "github.com/Masterminds/squirrel"
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

func (p *UpdateProductBlueprint) Exec(
	ctx context.Context,
	base pgx.Tx,
) (err error) {
	q := util.SQL

	fields := make(map[string]any)

	/*
	     	Name        *string          `form:"name"`
	   	Price       *string          `form:"price"`
	   	Stock       *int             `form:"stock"`
	   	Discount    *decimal.Decimal `form:"discount"`
	   	Description *string          `form:"description"`
	   	Category    *string          `form:"category"`
	   	Featured    *bool            `form:"featured"`

	   	Image      *multipart.FileHeader   `form:"imageUrl"`
	   	DescImages []*multipart.FileHeader `form:"descriptionImages"`
	   	SpecImages []*multipart.FileHeader `form:"specificationImages"`
	   	PackImages []*multipart.FileHeader `form:"deliveryImages"`

	   	SpecificationsJSON *string `form:"specifications"`
	   	ColorListJSON        *string `form:"colors"`

	   	CustomName           *bool   `form:"allowCustomName"`
	   	CustomModality       *bool   `form:"allowCustomModality"`
	   	CustomColorSelection *bool   `form:"allowCustomColorSelection"`
	*/

	update := q.Update("products")

	if p.Name != nil {
		fields["name"] = *p.Name
	}

	if p.Price != nil {
		fields["price"] = *p.Price
	}

	if p.Stock != nil {
		fields["stock"] = *p.Stock
	}

	if p.Discount != nil {
		fields["discount"] = *p.Discount
	}

	if p.Description != nil {
		fields["description"] = *p.Description
	}

	if p.Category != nil {
		var category string
		err = base.QueryRow(
			ctx,
			"select id from categories where name = $1",
			*p.Category,
		).Scan(&category)
		if err != nil {
			return fmt.Errorf("Error finding category: %w", err)
		}
		fields["category_id"] = category
	}

	if p.Featured != nil {
		fields["featured"] = *p.Featured
	}

	attrExpr := "attr"
	var attrArgs []any

	addSet := func(key string, val any) {
		attrExpr = fmt.Sprintf(query.JSONBSetFormat, attrExpr, key)
		attrArgs = append(attrArgs, val)
	}

	if p.ImageURL != nil {
		addSet("main_image", *p.ImageURL)
	}
	if p.CustomName != nil {
		addSet("customName", *p.CustomName)
	}
	if p.CustomModality != nil {
		addSet("customModality", *p.CustomModality)
	}
	if p.CustomColorSelection != nil {
		addSet("customColor", *p.CustomColorSelection)
	}

	if len(p.DescImagesURLs) > 0 {
		addSet("desc_images", p.DescImagesURLs)
	}
	if len(p.SpecImagesURLs) > 0 {
		addSet("spec_images", p.SpecImagesURLs)
	}
	if len(p.PackImagesURLs) > 0 {
		addSet("pack_images", p.PackImagesURLs)
	}

	if p.SpecificationsJSON != nil {
		addSet("specifications", *p.SpecificationsJSON)
	}
	if p.ColorListJSON != nil {
		addSet("colors", *p.ColorListJSON)
	}

	if len(attrArgs) > 0 {
		update = update.Set("attr", sq.Expr(attrExpr, attrArgs...))
	} else if len(fields) == 0 {
		return
	}

	update = update.
		SetMap(fields).
		Where(sq.Eq{"id": p.ID})

	query, params, err := update.ToSql()

	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	_, err = base.Exec(ctx, query, params...)

	if err != nil {
		return fmt.Errorf(
			"could not update consignment, query: %q, %w", query, err)
	}

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
