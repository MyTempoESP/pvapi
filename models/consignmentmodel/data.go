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

// ConsignmentWithShipment represents a row from the unified view (consignments + shipments).
// All fields are nullable where applicable (shipments have NULL comission/consignee/pickup).
type ConsignmentWithShipment struct {
	ID             int              `db:"consignment_id"`
	ComissionType  *string          `db:"consignment_comission_type"`
	ComissionValue *decimal.Decimal `db:"consignment_comission_value"`
	Status         string           `db:"consignment_status"`
	CreatedAt      time.Time        `db:"consignment_created_at"`
	Slug           string           `db:"consignment_slug"`
	Code           string           `db:"consignment_code"`

	Shop

	ConsigneeID   *int    `db:"consignee_id"`
	ConsigneeName *string `db:"consignee_name"`

	PickupID   *int    `db:"pickup_location_id"`
	PickupName *string `db:"pickup_location_name"`

	RecordType string `db:"record_type"` // "consignment" or "shipment"
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
