package clientmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

type ClientResource struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	TotalPurchases decimal.Decimal `json:"total_purchases"`
	LastPurchase   *time.Time      `json:"last_purchase,omitempty"`
}

func (c Client) ToResource() ClientResource {
	phone := ""
	if c.Phone != nil {
		phone = *c.Phone
	}
	return ClientResource{
		ID:             c.ID,
		Name:           c.Name,
		Email:          c.Email,
		Phone:          phone,
		TotalPurchases: c.TotalPurchases,
		LastPurchase:   c.LastPurchase,
	}
}
