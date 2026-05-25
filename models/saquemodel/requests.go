package saquemodel

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type CreateSaqueRequest struct {
	Description     string           `json:"description" binding:"required"`
	AskedValue      decimal.Decimal  `json:"asked_value" binding:"required"`
	BankName        string           `json:"bank_name" binding:"required"`
	PixCPF          *string          `json:"pix_cpf"`
	PixTitular      *string          `json:"pix_titular"`
	TransferDetails *json.RawMessage `json:"transfer_details"`
}

type UpdateSaqueRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected cancelled"`
}
