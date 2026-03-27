package consignmentmodel

import (
	"fmt"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/util"
)

type CreateConsignmentRequest struct {
	PickupID    int `json:"destinationId" binding:"required"`
	ConsigneeID int `json:"monitorId" binding:"required"`

	CutPCT   *decimal.Decimal `json:"monitorProfitPercentage"`
	CutFixed *decimal.Decimal `json:"monitorProfit"`

	Status string `json:"status" binding:"required"`
}

// TODO
type UpdateConsignmentRequest struct{}

func (c CreateConsignmentRequest) Prepare(shopID int) (ConsignmentBlueprint, error) {
	var zero ConsignmentBlueprint

	var comissionType string
	var comissionValue decimal.Decimal

	if c.CutPCT != nil {
		comissionType = "PERCENT"
		comissionValue = *c.CutPCT
	}

	if c.CutFixed != nil {
		comissionType = "FIXED"
		comissionValue = *c.CutFixed
	}

	// if still empty, we're kinda cooked
	if comissionType == "" {
		return zero, fmt.Errorf(
			"nenhum tipo de comissão especificado %w",
			util.BadRequestError,
		)
	}

	return ConsignmentBlueprint{
		ShopID:         shopID,
		PickupID:       c.PickupID,
		ConsigneeID:    c.ConsigneeID,
		ComissionType:  comissionType,
		ComissionValue: comissionValue,
		Status:         c.Status,
	}, nil
}

// TODO
func (c UpdateConsignmentRequest) Prepare(id int) Consignment {
	return Consignment{
		ID: id,
	}
}
