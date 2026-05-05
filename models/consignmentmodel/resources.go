package consignmentmodel

import (
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

// ConsignmentWithShipmentResource is the JSON representation for both consignments and shipments.
type ConsignmentWithShipmentResource struct {
	ID             int              `json:"id"`
	ComissionType  *string          `json:"commission_type,omitempty"`
	ComissionValue *decimal.Decimal `json:"commission_value,omitempty"`
	Status         string           `json:"status"`
	Code           string           `json:"code"`
	Slug           string           `json:"slug"`
	RecordType     string           `json:"record_type"`

	ShopName string `json:"shop_name"`
	ShopID   int    `json:"shop_id"`

	// Only present for consignments
	Consignee   *ConsigneeResource `json:"consignee,omitempty"`
	// For consignments: pickup location name; for shipments: delivery address text
	Destination *string `json:"destination,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type ConsigneeResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c ConsignmentWithShipment) ToResource() ConsignmentWithShipmentResource {
	res := ConsignmentWithShipmentResource{
		ID:             c.ID,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		Status:         c.Status,
		Code:           c.Code,
		Slug:           c.Slug,
		RecordType:     c.RecordType,
		ShopID:         c.Shop.ID,
		ShopName:       c.Shop.Name,
		CreatedAt:      c.CreatedAt,
	}

	if c.RecordType == "consignment" {
		if c.ConsigneeID != nil && c.ConsigneeName != nil {
			res.Consignee = &ConsigneeResource{
				ID:   *c.ConsigneeID,
				Name: *c.ConsigneeName,
			}
		}
		if c.PickupName != nil {
			res.Destination = c.PickupName
		}
	} else { // shipment
		// Use PickupName as the delivery address
		res.Destination = c.PickupName
		// No comission or consignee for shipments
		res.ComissionType = nil
		res.ComissionValue = nil
	}

	return res
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
		Image:            p.Product.Attributes.Image,
		OriginalPrice:    p.Product.Price,
		OriginalQuantity: p.Product.Stock,

		PackagePrice:    p.Package.Price,
		PackageQuantity: p.Package.Quantity,

		ProductResource: p.Product.ToResource(),

		// The nested "pivot" object
		Package: PackageResource{
			Price:    p.Package.Price,
			Quantity: p.Package.Quantity,
		},
	}
}
