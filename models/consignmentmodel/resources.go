package consignmentmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

type ProductResource struct {
	ID               int             `json:"id"`
	UUID             string          `json:"uuid"`
	Name             string          `json:"name"`
	Code             string          `json:"code"`
	Description      string          `json:"description"`
	Image            string          `json:"image"`
	CategoryID       int             `json:"category_id"`
	ShopID           int             `json:"shop_id"`
	Featured         bool            `json:"featured"`
	Discount         decimal.Decimal `json:"discount"`
	OriginalPrice    decimal.Decimal `json:"original_price"`
	OriginalQuantity int             `json:"original_quantity"`

	// duplication here is intentional for backwards compatibility
	PackagePrice    decimal.Decimal `json:"consignment_price"`
	PackageQuantity int             `json:"consignment_quantity"`

	Package PackageResource `json:"pivot"`
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

func (p Product) ToResource() ProductResource {
	return ProductResource{
		ID:               p.ID,
		UUID:             p.UUID,
		Name:             p.Name,
		Code:             p.Code,
		Description:      p.Description,
		Image:            p.Image,
		CategoryID:       p.CategoryID,
		ShopID:           p.ShopID,
		Featured:         p.Featured,
		Discount:         p.Discount,
		OriginalPrice:    p.OriginalPrice,
		OriginalQuantity: p.OriginalQuantity,

		PackagePrice:    p.Package.Price,
		PackageQuantity: p.Package.Quantity,

		// The nested "pivot" object
		Package: PackageResource{
			Price:    p.Package.Price,
			Quantity: p.Package.Quantity,
		},
	}
}
