package consignmentmodel

import "github.com/shopspring/decimal"

type ConsignmentBlueprint struct {
	ShopID      int
	PickupID    int
	ConsigneeID int

	ComissionType  string
	ComissionValue decimal.Decimal
}
