package consignmentmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

type ConsignmentProduct struct {
	// Base Product Fields
	ID               int             `db:"product_id"`
	Name             string          `db:"product_name"`
	OriginalPrice    decimal.Decimal `db:"product_price"`
	OriginalQuantity int             `db:"product_stock"`
	Discount         decimal.Decimal `db:"product_discount"`
	Featured         bool            `db:"product_featured"`
	Description      string          `db:"product_description"`
	Image            string          `db:"product_main_image"`
	CategoryID       int             `db:"product_category_id"`
	ShopID           int             `db:"product_shop_id"`
	Code             string          `db:"product_code"`
	UUID             string          `db:"product_uuid"`

	Package
}

type Package struct {
	Price    decimal.Decimal `db:"package_price"`
	Quantity int             `db:"package_quantity"`
}

type Consignee struct {
	ID   int    `db:"consignee_id"`
	Name string `db:"consignee_name"`
}

type Shop struct {
	ID   int    `db:"shop_id"`
	Name string `db:"shop_name"`
}

type Pickup struct {
	ID   int    `db:"pickup_location_id"`
	Name string `db:"pickup_location_name"`
}

type Consignment struct {
	ID             int             `db:"consignment_id"`
	ComissionType  string          `db:"consignment_comission_type"`
	ComissionValue decimal.Decimal `db:"consignment_comission_value"`
	Status         string          `db:"consignment_status"`
	CreatedAt      time.Time       `db:"consignment_created_at"`
	Slug           string          `db:"consignment_slug"`
	Code           string          `db:"consignment_code"`

	Shop
	Consignee
	Pickup
}

type DetailedConsignment struct {
	Consignment
}
