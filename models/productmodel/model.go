package productmodel

import (
	"encoding/json"
	"context"
	"slices"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/models/model"
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

func generateProductAttr(
	ctx context.Context,
	base pgx.Tx,
	p *UpdateProductBlueprint,
) (attr ProductAttr, err error) {
	err = base.QueryRow(ctx, query.ProductGetAttrQuery, p.ID).Scan(&attr)

	if err != nil {
		return
	}
	
	// Strings
	if p.ImageURL != nil {
		attr.Image = *p.ImageURL
	}

	// Bools
	if p.CustomName != nil {
		attr.AllowCustomName = *p.CustomName
	}
	if p.CustomModality != nil {
		attr.AllowCustomModality = *p.CustomModality
	}
	if p.CustomColorSelection != nil {
		attr.AllowCustomColor = *p.CustomColorSelection
	}

	// Lists
	if len(p.DescImagesURLs) > 0 {
		attr.DescImages = slices.Concat(attr.DescImages, p.DescImagesURLs)
	}
	if len(p.SpecImagesURLs) > 0 {
		attr.SpecImages = slices.Concat(attr.SpecImages, p.SpecImagesURLs)
	}
	if len(p.PackImagesURLs) > 0 {
		attr.PackImages = slices.Concat(attr.PackImages, p.PackImagesURLs)
	}

	// Objects
	if p.SpecificationsJSON != nil {
		err = json.Unmarshal([]byte(*p.SpecificationsJSON), &attr.Specifications)

		if err != nil {
			return
		}
	}
	if p.ColorListJSON != nil {
		err = json.Unmarshal([]byte(*p.ColorListJSON), &attr.Colors)

		if err != nil {
			return
		}
	}

	return
}

func (p *UpdateProductBlueprint) Exec(
	ctx context.Context,
	base pgx.Tx,
) (err error) {
	attr, err := generateProductAttr(ctx, base, p)

	if err != nil {
		return
	}

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

	fields["attr"] = attr

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

func HasImage(
	ctx context.Context,
	base db.Queryable,
	productID int,
	imageUrl string,
) (bool, error) {
	return model.HasResource(
		ctx, base,
		query.ProductHasImage,
		productID, imageUrl,
	)
}

func DeleteImage(
	ctx context.Context,
	tx pgx.Tx,
	productID int,
	imageUrl string,
) error {
	_, err := tx.Exec(
		ctx,
		query.ProductDeleteImageQuery,
		productID, imageUrl,
	)

	return err
}

