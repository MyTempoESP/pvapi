package saquemodel

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type PixResource struct {
	Key     string `json:"key"`
	Titular string `json:"titular"`
}

type SaqueResource struct {
	ID              int              `json:"id"`
	Description     string           `json:"description"`
	AskedValue      decimal.Decimal  `json:"asked_value"`
	UserName        string           `json:"user_name"`
	ShopName        string           `json:"shop_name"`
	BankName        string           `json:"bank_name"`
	Pix             *PixResource     `json:"pix,omitempty"`
	TransferDetails *json.RawMessage `json:"transfer_details,omitempty"`
	Status          string           `json:"status"`
	Motivo          string           `json:"motivo"`
	CreatedAt       string           `json:"created_at"`
}

func (r SaqueRow) ToResource() SaqueResource {
	sr := SaqueResource{
		ID:          r.ID,
		Description: r.Description,
		AskedValue:  r.AskedValue,
		UserName:    r.UserName,
		ShopName:    r.ShopName,
		Status:      r.Status,
		Motivo:      r.Motivo,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
	}

	if r.BankName != nil {
		sr.BankName = *r.BankName
	}

	if r.PixKey != nil || r.PixTitular != nil {
		sr.Pix = &PixResource{}
		if r.PixKey != nil {
			sr.Pix.Key = *r.PixKey
		}
		if r.PixTitular != nil {
			sr.Pix.Titular = *r.PixTitular
		}
	}

	if len(r.TransferDetails) > 0 && string(r.TransferDetails) != "null" {
		details := json.RawMessage(r.TransferDetails)
		sr.TransferDetails = &details
	}

	return sr
}
