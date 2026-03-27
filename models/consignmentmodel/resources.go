package consignmentmodel

import "github.com/shopspring/decimal"

type ShopResource struct {
}

type ProductResource struct {
}

type ConsignmentResource struct {
	ID             int             `json:"id"`
	ComissionType  string          `json:"commission_type"`
	ComissionValue decimal.Decimal `json:"commission_value"`
	Status         string          `json:"status"`

	Pickup int `json:"pickup_location_id"`

	ShopID   int    `json:"establishment_id"`
	ShopName string `json:"establishment_name"`

	ConsigneeID   int    `json:"monitor_id"`
	ConsigneeName string `json:"monitor_name"`

	Items []ProductResource `json:"items"`
}

func (c Consignment) ToResource() ConsignmentResource {
	return ConsignmentResource{
		ID:             c.ID,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		Status:         c.Status,

		Pickup: c.Pickup,

		ShopID:   c.Shop.ID,
		ShopName: c.Shop.Name,

		ConsigneeID:   c.Consignee.ID,
		ConsigneeName: c.Consignee.Name,

		Items: []ProductResource{},
	}
}
