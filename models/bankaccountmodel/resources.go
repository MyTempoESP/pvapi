package bankaccountmodel

import (
	"encoding/json"
	"time"

	"linhazero.com/pvapi/models/saquemodel"
)

type BankAccountResource struct {
	ID              int                    `json:"id"`
	ShopID          int                    `json:"shop_id"`
	Name            string                 `json:"name"`
	BankName        string                 `json:"bank_name"`
	Pix             *saquemodel.PixResource `json:"pix,omitempty"`
	TransferDetails *json.RawMessage       `json:"transfer_details,omitempty"`
	CreatedAt       string                 `json:"created_at"`
}

func (r BankAccountRow) ToResource() BankAccountResource {
	br := BankAccountResource{
		ID:       r.ID,
		ShopID:   r.ShopID,
		Name:     r.Name,
		BankName: r.BankName,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
	}

	if r.PixKey != nil || r.PixTitular != nil {
		br.Pix = &saquemodel.PixResource{}
		if r.PixKey != nil {
			br.Pix.Key = *r.PixKey
		}
		if r.PixTitular != nil {
			br.Pix.Titular = *r.PixTitular
		}
	}

	if len(r.TransferDetails) > 0 && string(r.TransferDetails) != "null" {
		details := json.RawMessage(r.TransferDetails)
		br.TransferDetails = &details
	}

	return br
}
