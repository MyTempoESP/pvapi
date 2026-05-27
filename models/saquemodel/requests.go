package saquemodel

import (
	"github.com/shopspring/decimal"
)

type CreateSaqueRequest struct {
	Description   string          `json:"description" binding:"required"`
	AskedValue    decimal.Decimal `json:"asked_value" binding:"required"`
	BankAccountID int             `json:"bank_account" binding:"required"`
}

type UpdateSaqueRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected cancelled"`
}
