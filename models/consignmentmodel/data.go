package consignmentmodel

import (
	"time"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/models/productmodel"
)

type ConsignmentProduct struct {
	productmodel.Product
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
	Name           string          `db:"consignment_name"`
	ComissionType  string          `db:"consignment_comission_type"`
	ComissionValue decimal.Decimal `db:"consignment_comission_value"`
	Status         string          `db:"consignment_status"`
	CreatedAt      time.Time       `db:"consignment_created_at"`
	Slug           string          `db:"consignment_slug"`
	Code           string          `db:"consignment_code"`

	Shop
	Consignee
	Pickups []Pickup `db:"-"`
}

type ConsignmentWithStats struct {
	ID                int             `db:"consignment_id"`
	Name              string          `db:"consignment_name"`
	ComissionType     string          `db:"consignment_comission_type"`
	ComissionValue    decimal.Decimal `db:"consignment_comission_value"`
	Status            string          `db:"consignment_status"`
	CreatedAt         time.Time       `db:"consignment_created_at"`
	Slug              string          `db:"consignment_slug"`
	Code              string          `db:"consignment_code"`

	TotalProducts     int `db:"total_products"`
	ConfirmedProducts int `db:"confirmed_products"`

	Shop
	Consignee
	Pickups []Pickup `db:"-"`
}

type DetailedConsignment struct {
	Consignment
}
