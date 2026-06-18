package pickupmodel

import (
	"time"

	"linhazero.com/pvapi/address"
)

type CreatePickupLocationRequest struct {
	Name     string                          `json:"name"`
	DateTime string                          `json:"datetime"`
	Address  *address.AddressRequestOptional `json:"address"`
}

func (c CreatePickupLocationRequest) Prepare(
	shopID int,
) (PickupLocationBlueprint, error) {
	var zero PickupLocationBlueprint

	t, err := time.Parse("2006-01-02T15:04:05", c.DateTime)
	if err != nil {
		return zero, err
	}

	addr := address.Address{Number: "0"}
	if c.Address != nil {
		addr, err = c.Address.Prepare()
		if err != nil {
			return zero, err
		}
	}

	return PickupLocationBlueprint{
		Name:     c.Name,
		Datetime: t,
		Address:  addr,
		ShopID:   shopID,
	}, nil
}

type UpdatePickupLocationRequest struct {
	Name     *string                          `json:"name"`
	DateTime *string                          `json:"datetime"`
	Address  *address.AddressRequestOptional `json:"address"`
}
