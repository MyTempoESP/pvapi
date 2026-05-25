package shopmodel

import (
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/models/saquemodel"
)

type CheckoutResource struct {
	Description string          `json:"description"`
	ID          string          `json:"id"`
	Value       decimal.Decimal `json:"value"`
	Tax         decimal.Decimal `json:"tax"`
	Commission  decimal.Decimal `json:"commission"`
}

type ShopCheckoutReport struct {
	Checkouts []CheckoutResource          `json:"checkouts"`
	Totals    TotalsResource              `json:"totals"`
	Saques    []saquemodel.SaqueResource  `json:"saques"`
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
