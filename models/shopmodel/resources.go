package shopmodel

import (
	"github.com/gosimple/slug"
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

type DetailedShopResource struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Telephone string `json:"phone"`
	Status    string `json:"status"`
	Products  int    `json:"products_count"`
	Pickups   int    `json:"pickup_locations_count"`

	Address address.Address `json:"address"`
	User    usermodel.User  `json:"owner"`
}

func (shop Shop) ToResource() ShopResource {
	return ShopResource{
		ID:        shop.ID,
		Name:      shop.Name,
		Slug:      shop.Slug,
		Telephone: shop.Telephone,
		Status:    shop.Status,

		Address: shop.Address,
		User:    shop.User,
	}
}

func (shop DetailedShop) ToResource() DetailedShopResource {
	return DetailedShopResource{
		ID:        shop.ID,
		Name:      shop.Name,
		Slug:      shop.Slug,
		Telephone: shop.Telephone,
		Status:    shop.Status,
		Products:  shop.Products,
		Pickups:   shop.Pickups,

		Address: shop.Address,
		User:    shop.User,
	}
}

func (shop Shop) Summarize() ShopSummaryResource {
	return ShopSummaryResource{
		ID:        shop.ID,
		Name:      shop.Name,
		Slug:      shop.Slug,
		Telephone: shop.Telephone,
	}
}

func (shop DetailedShop) Summarize() ShopSummaryResource {
	return ShopSummaryResource{
		ID:        shop.ID,
		Name:      shop.Name,
		Slug:      shop.Slug,
		Telephone: shop.Telephone,
	}
}

func SummarizeRequest(shopID int, shop CreateShopRequest) ShopSummaryResource {
	return ShopSummaryResource{
		ID:        shopID,
		Name:      shop.Name,
		Slug:      slug.Make(shop.Name),
		Telephone: shop.Telephone,
	}
}
