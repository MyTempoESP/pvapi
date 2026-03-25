package consigneemodel

import (
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/util/userutils"
)

type ConsigneeBlueprint struct {
	Name   string
	Email  string
	Phone  string
	ShopID int
	User   userutils.User

	Address address.AddressRequest
}

func (c CreateConsigneeRequest) Prepare(shop int) (*ConsigneeBlueprint, error) {

	user, err := userutils.NewUser(
		c.Name,
		c.Email,
		c.Phone,
	)

	if err != nil {
		return nil, err
	}

	return &ConsigneeBlueprint{
		Name:   c.Name,
		Email:  c.Email,
		Phone:  c.Phone,
		ShopID: shop,
		User:   user,

		Address: *c.Address,
	}, nil
}
