package shopmodel

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CheckoutResource struct {
	Description string          `json:"description"`
	ID          string          `json:"id"`
	Value       decimal.Decimal `json:"value"`
	Tax         decimal.Decimal `json:"tax"`
	Commission  decimal.Decimal `json:"commission"`
}

type SaqueResource struct {
	Description     string           `json:"description"`
	AskedValue      decimal.Decimal  `json:"asked_value"`
	UserName        string           `json:"user_name"`
	ShopName        string           `json:"shop_name"`
	BankName        string           `json:"bank_name"`
	Pix             *PixResource     `json:"pix,omitempty"`
	TransferDetails *json.RawMessage `json:"transfer_details,omitempty"`
	CreatedAt       string           `json:"created_at"`
	ID              int              `json:"id"`
	Status          string           `json:"status"`
}

type PixResource struct {
	CPF     string `json:"cpf"`
	Titular string `json:"titular"`
}

type ShopCheckoutReport struct {
	Checkouts []CheckoutResource `json:"checkouts"`
	Totals    TotalsResource     `json:"totals"`
	Saques    []SaqueResource    `json:"saques"`
}

type TotalsResource struct {
	Taxes       decimal.Decimal `json:"taxes"`
	Commissions decimal.Decimal `json:"commissions"`
	Total       decimal.Decimal `json:"total"`
	Saques      decimal.Decimal `json:"saques"`
	Actual      decimal.Decimal `json:"actual"`
}

func (r ShopCheckoutRow) ToResource() CheckoutResource {
	return CheckoutResource{
		Description: r.Description,
		ID:          r.UniqueID,
		Value:       r.Value,
		Tax:         r.Tax,
		Commission:  r.Commission,
	}
}

func (r SaqueRow) ToResource() SaqueResource {
	sr := SaqueResource{
		Description: r.Description,
		AskedValue:  r.AskedValue,
		UserName:    r.UserName,
		ShopName:    r.ShopName,
		BankName:    r.BankName,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		ID:          r.ID,
		Status:      r.Status,
	}

	if r.PixCPF != nil || r.PixTitular != nil {
		sr.Pix = &PixResource{}
		if r.PixCPF != nil {
			sr.Pix.CPF = *r.PixCPF
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
