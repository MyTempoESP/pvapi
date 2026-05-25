package consignmentmodel

import (
	"fmt"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/util"
)

type AddProductUnit struct {
	ID          int             `json:"productId" binding:"required"`
	Quantity    int             `json:"quantity"  binding:"required"`
	Price       decimal.Decimal `json:"price"     binding:"required"`
	Color       string          `json:"color"     binding:"required"`
	Discount    decimal.Decimal `json:"discount"  binding:"required"`
	Featured    bool            `json:"featured"  binding:"required"`
	Name        string          `json:"productName"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
}

type AddProductRequest struct {
	Products []AddProductUnit `json:"products"`
}

type Comission struct {
	CutPCT   *decimal.Decimal `json:"monitorProfitPercentage"`
	CutFixed *decimal.Decimal `json:"monitorProfit"`
}

type CreateConsignmentRequest struct {
	PickupID    int    `json:"destinationId" binding:"required"`
	ConsigneeID int    `json:"monitorId" binding:"required"`
	Status      string `json:"status" binding:"required"`

	Comission
}

type UpdateConsignmentRequest struct {
	PickupID    *int    `json:"destinationId"`
	ConsigneeID *int    `json:"monitorId"`
	Status      *string `json:"status"`

	Comission
}

type UpdateConsignmentProductRequest struct {
	Price    *decimal.Decimal `json:"price"`
	Quantity *int             `json:"quantity"`
}

func (p UpdateConsignmentProductRequest) Prepare(
	consignmentID, productID int,
) UpdateConsignmentProductBlueprint {
	return UpdateConsignmentProductBlueprint{
		ProductID:                       productID,
		ConsignmentID:                   consignmentID,
		UpdateConsignmentProductRequest: p,
	}
}

func (c Comission) defineComissionType() (
	comissionType string,
	comissionValue decimal.Decimal,
	err error,
) {
	if c.CutPCT != nil {
		comissionType = ComissionPercent
		comissionValue = *c.CutPCT
	}

	if c.CutFixed != nil {
		comissionType = ComissionFixed
		comissionValue = *c.CutFixed
	}

	// if still empty, we're kinda cooked
	if comissionType == "" {
		err = fmt.Errorf(
			"nenhum tipo de comissão especificado %w",
			util.ErrBadRequest,
		)
		return
	}

	return
}

func (c CreateConsignmentRequest) Prepare(shopID int) (ConsignmentBlueprint, error) {
	var zero ConsignmentBlueprint

	comissionType, comissionValue, err := c.defineComissionType()

	if err != nil {
		return zero, err
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

func (c UpdateConsignmentRequest) Prepare(id int) ConsignmentUpdateBlueprint {
	var comissionType *string = nil
	var comissionValue *decimal.Decimal = nil

	t, v, err := c.defineComissionType()

	if err == nil {
		comissionType = &t
		comissionValue = &v
	}

	return ConsignmentUpdateBlueprint{
		ID:             id,
		PickupID:       c.PickupID,
		ConsigneeID:    c.ConsigneeID,
		ComissionType:  comissionType,
		ComissionValue: comissionValue,
		Status:         c.Status,
	}
}
