package pickupmodel

import (
	"time"

	"linhazero.com/pvapi/address"
)

type CreatePickupLocationRequest struct {
	Name     string `json:"name"`
	DateTime string `json:"datetime"`
	Address  *address.Address
}

func (c CreatePickupLocationRequest) Prepare(
	shopID int,
) (PickupLocationBlueprint, error) {
	var zero PickupLocationBlueprint

	t, err := time.Parse(time.RFC3339, c.DateTime)

	if err != nil {
		return zero, nil
	}

	return PickupLocationBlueprint{
		Name:     c.Name,
		Datetime: t,
		Address:  c.Address,
		ShopID:   shopID,
	}, nil
}
