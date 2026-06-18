package consignmentmodel

import "github.com/shopspring/decimal"

type ConsignmentBlueprint struct {
	ShopID      int
	PickupIDs   []int
	ConsigneeID int

	ComissionType  string
	ComissionValue decimal.Decimal

	Name   string
	Status string
}

type ConsignmentUpdateBlueprint struct {
	ID             int
	ConsigneeID    *int
	ComissionType  *string
	ComissionValue *decimal.Decimal
	Name           *string
	Status         *string
}

type UpdateConsignmentProductBlueprint struct {
	ProductID     int
	ConsignmentID int

	UpdateConsignmentProductRequest
}
