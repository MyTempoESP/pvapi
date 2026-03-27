package consignmentmodel

import "github.com/shopspring/decimal"

type ConsignmentResource struct {
	ConsigneeName  string          `json:"monitor_name"`
	ShopName       string          `json:"establishment_name"`
	ComissionType  string          `json:"commission_type"`
	ComissionValue decimal.Decimal `json:"commission_value"`
	PickupName     string          `json:"pickup_point_name"`
}

func (c Consignment) ToResource() ConsignmentResource {
	return ConsignmentResource{
		ConsigneeName:  c.ConsigneeName,
		ShopName:       c.ShopName,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		PickupName:     c.PickupName,
	}
}
