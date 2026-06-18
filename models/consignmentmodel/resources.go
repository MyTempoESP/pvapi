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

type PackageResource struct {
	Price    decimal.Decimal `json:"price"`
	Quantity int             `json:"quantity"`
}

type DestinationsResource []PickupResource

type PickupResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewPickupResourceFromPickup(p Pickup) PickupResource {
	return PickupResource(p)
}

type ConsignmentResource struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	ComissionType  string          `json:"commission_type"`
	ComissionValue decimal.Decimal `json:"commission_value"`
	Code           string          `json:"code"`
	Slug           string          `json:"slug"`
	Status         string          `json:"status"`

	Destinations  DestinationsResource `json:"destinations"`
	ConsigneeName string               `json:"responsibleMonitor"`

	CreatedAt time.Time `json:"created_at"`
}

type ConsignmentWithStatsResource struct {
	Code              string               `json:"code"`
	ID                int                  `json:"id"`
	Name              string               `json:"name"`
	Status            string               `json:"status"`
	Destinations      DestinationsResource `json:"destinations"`
	MonitorName       string               `json:"monitorName"`
	CreatedAt         time.Time            `json:"createdAt"`
	TotalProducts     int                  `json:"totalProducts"`
	ConfirmedProducts int                  `json:"confirmedProducts"`
}

func (c ConsignmentWithStats) ToResource() ConsignmentWithStatsResource {
	return ConsignmentWithStatsResource{
		Code:              c.Code,
		ID:                c.ID,
		Name:              c.Name,
		Status:            c.Status,
		Destinations:      pickupsToDestinations(c.Pickups),
		MonitorName:       c.Consignee.Name,
		CreatedAt:         c.CreatedAt,
		TotalProducts:     c.TotalProducts,
		ConfirmedProducts: c.ConfirmedProducts,
	}
}

type PublicConsignmentResource struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	ComissionType  string          `json:"commission_type"`
	ComissionValue decimal.Decimal `json:"commission_value"`
	Code           string          `json:"code"`
	Slug           string          `json:"slug"`
	Status         string          `json:"status"`

	Destinations  DestinationsResource `json:"destinations"`
	ConsigneeName string               `json:"responsibleMonitor"`

	Shop struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"shop"`

	CreatedAt time.Time `json:"created_at"`
}

func pickupsToDestinations(pickups []Pickup) DestinationsResource {
	if pickups == nil {
		return DestinationsResource{}
	}
	res := make(DestinationsResource, len(pickups))
	for i, p := range pickups {
		res[i] = NewPickupResourceFromPickup(p)
	}
	return res
}

func (c Consignment) ToResource() ConsignmentResource {
	return ConsignmentResource{
		ID:             c.ID,
		Name:           c.Name,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		Code:           c.Code,
		Slug:           c.Slug,
		Status:         c.Status,

		Destinations:  pickupsToDestinations(c.Pickups),
		ConsigneeName: c.Consignee.Name,

		CreatedAt: c.CreatedAt,
	}
}

func (c Consignment) ToPublicResource() PublicConsignmentResource {
	r := PublicConsignmentResource{
		ID:             c.ID,
		Name:           c.Name,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		Code:           c.Code,
		Slug:           c.Slug,
		Status:         c.Status,

		Destinations:  pickupsToDestinations(c.Pickups),
		ConsigneeName: c.Consignee.Name,

		CreatedAt: c.CreatedAt,
	}
	r.Shop.ID = c.Shop.ID
	r.Shop.Name = c.Shop.Name
	return r
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
