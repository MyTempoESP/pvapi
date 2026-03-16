package shopmodel

import (
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/models/usermodel"
)

type ShopSummaryResource struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Telephone string `json:"phone"`
	Status    string `json:"status"`
}

type ShopResource struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Telephone string `json:"phone"`
	Status    string `json:"status"`

	Address address.Address `json:"address"`
	User    usermodel.User  `json:"owner"`
}
