package bankaccountmodel

import "encoding/json"

type CreateBankAccountRequest struct {
	Name            string           `json:"name" binding:"required"`
	BankName        string           `json:"bank_name" binding:"required"`
	PixKey          *string          `json:"pix_key"`
	PixTitular      *string          `json:"pix_titular"`
	TransferDetails *json.RawMessage `json:"transfer_details"`
}

type UpdateBankAccountRequest struct {
	Name            string           `json:"name"`
	BankName        string           `json:"bank_name"`
	PixKey          *string          `json:"pix_key"`
	PixTitular      *string          `json:"pix_titular"`
	TransferDetails *json.RawMessage `json:"transfer_details"`
}
