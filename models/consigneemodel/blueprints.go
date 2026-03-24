package consigneemodel

import "linhazero.com/pvapi/address"

type ConsigneeBlueprint struct {
	Name    string
	Email   string
	Phone   string
	ShopID  int
	Address address.AddressRequest
}

func (c CreateConsigneeRequest) Prepare(shop int) *ConsigneeBlueprint {
	return &ConsigneeBlueprint{
		Name:    c.Name,
		Email:   c.Email,
		Phone:   c.Phone,
		ShopID:  shop,
		Address: *c.Address,
	}
}
