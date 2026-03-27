package consignmentmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

type ShopResource struct {
}

type ProductResource struct {
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
