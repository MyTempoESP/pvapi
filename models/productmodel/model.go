package productmodel

import (
	"encoding/json"
	"context"
	"slices"
	"fmt"
	"strings"

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
	CategoryID       int             `db:"product_category_id"`
	Category         string          `db:"product_category"`
	UUID             string          `db:"product_uuid"`
	Code             string          `db:"product_code"`
	Attributes       ProductAttr     `db:"product_attr"`
	PlatformCutType  string          `db:"product_platform_cut_type"`
	PlatformCut      decimal.Decimal `db:"product_platform_cut"`
	Pickups          []ProductPickupResource `db:"-"`
}

func PublicIndex(
	ctx context.Context,
	base db.Queryable,
	shopIDs, consignmentIDs []int,
) ([]Product, error) {
	if len(consignmentIDs) > 0 {
		return publicIndexWithFilters(ctx, base, shopIDs, consignmentIDs)
	}

	if len(shopIDs) > 0 {
		q := util.SQL.Select(
			"product_id",
			"product_name",
			"product_price",
			"product_stock",
			"product_discount",
			"product_featured",
			"product_moderation_status",
			"product_description",
			"product_category_id",
			"product_category",
			"product_uuid",
			"product_code",
			"product_attr",
			"product_platform_cut",
			"product_platform_cut_type",
			"product_shop_id",
			"product_shop_name",
		).From("product_catalog_view").
			Where("product_moderation_status = 'approved'").
			Where(sq.Eq{"product_shop_id": shopIDs})

		sql, args, err := q.ToSql()
		if err != nil {
			return nil, fmt.Errorf("build shop-filter query: %w", err)
		}

		return util.WithCollect[Product](ctx, base, sql, args...)
	}

	return util.WithCollect[Product](
		ctx, base,
		query.GetAllPublicProducts,
	)
}

func publicIndexWithFilters(
	ctx context.Context,
	base db.Queryable,
	shopIDs, consignmentIDs []int,
) ([]Product, error) {
	cids := make([]any, len(consignmentIDs))
	for i, id := range consignmentIDs {
		cids[i] = id
	}

	cPlaceholders := make([]string, len(consignmentIDs))
	for i := range consignmentIDs {
		cPlaceholders[i] = "?"
	}
	cList := strings.Join(cPlaceholders, ", ")

	consignmentWhere := fmt.Sprintf(
		`EXISTS (SELECT 1 FROM packages p`+
			` JOIN consignments_active ca ON ca.id = p.consignment_id`+
			` WHERE p.product_id = v.product_id`+
			` AND p.consignment_id IN (%s))`,
		cList,
	)

	q := util.SQL.Select(
		"v.product_id",
		"v.product_name",
		"v.product_price",
		"v.product_stock",
		"v.product_discount",
		"v.product_featured",
		"v.product_moderation_status",
		"v.product_description",
		"v.product_category_id",
		"v.product_category",
		"v.product_uuid",
		"v.product_code",
		"v.product_attr",
		"v.product_platform_cut",
		"v.product_platform_cut_type",
		"v.product_shop_id",
		"v.product_shop_name",
	).
		From("product_catalog_view v").
		Where("v.product_moderation_status = 'approved'").
		Where(sq.Expr(consignmentWhere, cids...))

	if len(shopIDs) > 0 {
		sids := make([]any, len(shopIDs))
		for i, id := range shopIDs {
			sids[i] = id
		}
		sPlaceholders := make([]string, len(shopIDs))
		for i := range shopIDs {
			sPlaceholders[i] = "?"
		}
		sList := strings.Join(sPlaceholders, ", ")

		q = q.Where(sq.Expr("v.product_shop_id IN ("+sList+")", sids...))
	}

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build consignment-filter query: %w", err)
	}

	return util.WithCollect[Product](ctx, base, sql, args...)
}

type packageRow struct {
	ProductID     int             `db:"product_id"`
	PackageID     int             `db:"package_id"`
	PackagePrice  decimal.Decimal `db:"package_price"`
	PackageStock  int             `db:"package_stock"`
	Discount      decimal.Decimal `db:"package_discount"`
	ConsignmentID int             `db:"consignment_id"`
	PickupID      int             `db:"pickup_id"`
	PickupName    string          `db:"pickup_name"`
	PickupDate    string          `db:"pickup_date"`
	AddrStreet    string          `db:"addrstreet"`
	AddrNumber    string          `db:"addrnumber"`
	AddrCity      string          `db:"addrcity"`
	AddrState     string          `db:"addrstate"`
	AddrCEP       string          `db:"addrcep"`
	AddrNeigh     string          `db:"addrneigh"`
	AddrComplement string         `db:"addrcomplement"`
}

func BatchFetchPickups(
	ctx context.Context,
	base db.Queryable,
	productIDs []int,
) (map[int][]ProductPickupResource, error) {
	if len(productIDs) == 0 {
		return map[int][]ProductPickupResource{}, nil
	}

	ids := make([]any, len(productIDs))
	for i, id := range productIDs {
		ids[i] = id
	}

	rows, err := util.WithCollect[packageRow](
		ctx, base,
		query.PackagesByProductIDs,
		ids,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[int][]ProductPickupResource, len(productIDs))
	for _, r := range rows {
		result[r.ProductID] = append(result[r.ProductID], ProductPickupResource{
			PackageID:  r.PackageID,
			Price:      r.PackagePrice,
			Discount:   r.Discount,
			Quantity:   r.PackageStock,
			PickupName: r.PickupName,
			PickupDate: r.PickupDate,
			Address: ProductPickupAddressResource{
				Street:     r.AddrStreet,
				Number:     r.AddrNumber,
				Neigh:      r.AddrNeigh,
				City:       r.AddrCity,
				State:      r.AddrState,
				CEP:        r.AddrCEP,
				Complement: r.AddrComplement,
			},
		})
	}

	// Initialize empty slices for products with no packages
	for _, id := range productIDs {
		if _, ok := result[id]; !ok {
			result[id] = []ProductPickupResource{}
		}
	}

	return result, nil
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
			"category_id": p.CategoryID,
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

	if p.CategoryID != nil {
		fields["category_id"] = *p.CategoryID
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
) (Product, error) {
	return util.WithCollectExactlyOneRow[Product](
		ctx,
		base,
		query.FindProductQuery,
		product,
	)
}

/*func Take(
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
}*/

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

func Delete(
	ctx context.Context,
	tx pgx.Tx,
	productID int,
) error {
	// TODO: worker script to sweep public/uploads
	// i'm too lazy to do cleanup here
	_, err := tx.Exec(ctx, query.DeleteProduct, productID)
	return err
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


func Take(
	ctx context.Context,
	tx pgx.Tx,
	productID int,
	newQty, oldQty int,
) (err error) {
	dx := newQty - oldQty

	// check if new quantity is actually significant
	if dx == 0 {
		return
	}

	// TODO: do negative qtys work?

	// use the difference so we don't double-decrement
	_, err = tx.Exec(
		ctx,
		query.ProductTakeFromStock,
		productID,
		dx,
	)

	return
}


func TakeFromPackage(
	ctx context.Context,
	tx pgx.Tx,
	productID int,
	consignmentID int,
	qty int,
) (err error) {
	var currentQty int

	// TODO: there can be multiple packages with
	// the same product, beware
	if err = tx.QueryRow(
		ctx,
		query.VerifyPackageQuantity,
		productID,
		consignmentID,
	).Scan(&currentQty); err != nil {
		return
	}

	return Take(ctx, tx, productID, qty, currentQty)
}
