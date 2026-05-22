package consignmentmodel

import (
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/models/productmodel"
)

type ProductResource struct {
	Image            string          `json:"image"`
	OriginalPrice    decimal.Decimal `json:"original_price"`
	OriginalQuantity int             `json:"original_quantity"`

	// duplication here is intentional for backwards compatibility
	PackagePrice    decimal.Decimal `json:"consignment_price"`
	PackageQuantity int             `json:"consignment_quantity"`

	productmodel.ProductResource
	Package PackageResource `json:"pivot"`
}

type ConsignmentWithShipmentResource struct {
	ID             string           `json:"id"`
	ComissionType  *string          `json:"commission_type,omitempty"`
	ComissionValue *decimal.Decimal `json:"commission_value,omitempty"`
	Status         string           `json:"status"`
	Code           string           `json:"code"`
	Slug           string           `json:"slug"`
	RecordType     string           `json:"record_type"`

	ShopName string `json:"shop_name"`
	ShopID   int    `json:"shop_id"`

	// present for consignments
	ConsigneeName string `json:"responsibleMonitor,omitempty"`

	// consignment: pickup location name; shipment: delivery address text
	Destination string `json:"destination,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

func (c ConsignmentWithShipment) ToResource() (
	r ConsignmentWithShipmentResource,
) {
	r = ConsignmentWithShipmentResource{
		Status:     c.Status,
		Code:       c.Code,
		Slug:       c.Slug,
		RecordType: c.RecordType,
		ShopID:     c.Shop.ID,
		ShopName:   c.Name,
		CreatedAt:  c.CreatedAt,
	}

	if c.RecordType == "consignment" {
		r.ID = strconv.Itoa(c.ID)

		r.ComissionType = c.ComissionType
		r.ComissionValue = c.ComissionValue

		if c.ConsigneeID != nil && c.ConsigneeName != nil {
			r.ConsigneeName = *c.ConsigneeName
		}

		if c.PickupName != nil {
			r.Destination = *c.PickupName
		}

		return
	}

	// Shipment
	r.ID = fmt.Sprintf("V-%d", c.ID)

	// generic name for consignee since shipments are based
	r.ConsigneeName = "sistema"

	// PickupName as the delivery address
	r.Destination = *c.PickupName

	return
}

type PackageResource struct {
	Price    decimal.Decimal `json:"price"`
	Quantity int             `json:"quantity"`
}

type ConsignmentResource struct {
	ID             int             `json:"id"`
	ComissionType  string          `json:"commission_type"`
	ComissionValue decimal.Decimal `json:"commission_value"`
	Code           string          `json:"code"`
	Slug           string          `json:"slug"`
	Status         string          `json:"status"`

	PickupName    string `json:"destination"`
	ConsigneeName string `json:"responsibleMonitor"`

	CreatedAt time.Time `json:"created_at"`
}

func (c Consignment) ToResource() ConsignmentResource {
	return ConsignmentResource{
		ID:             c.ID,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		Code:           c.Code,
		Slug:           c.Slug,
		Status:         c.Status,

		PickupName:    c.Pickup.Name,
		ConsigneeName: c.Consignee.Name,

		CreatedAt: c.CreatedAt,
	}
}

func (p ConsignmentProduct) ToResource() ProductResource {
	return ProductResource{
		Image:            p.Attributes.Image,
		OriginalPrice:    p.Product.Price,
		OriginalQuantity: p.Stock,

		PackagePrice:    p.Package.Price,
		PackageQuantity: p.Quantity,

		ProductResource: p.Product.ToResource(),

		// The nested "pivot" object
		Package: PackageResource{
			Price:    p.Package.Price,
			Quantity: p.Quantity,
		},
	}
}
