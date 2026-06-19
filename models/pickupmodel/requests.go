package pickupmodel

import (
	"fmt"
	"time"

	"linhazero.com/pvapi/address"
)

type CreatePickupLocationRequest struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	DateTime    string                          `json:"datetime"`
	ValidUntil  string                          `json:"validUntil"`
	Address     *address.AddressRequestOptional `json:"address"`
}

func (c CreatePickupLocationRequest) Prepare(
	shopID int,
) (PickupLocationBlueprint, error) {
	var zero PickupLocationBlueprint

	t, err := time.Parse("2006-01-02T15:04:05", c.DateTime)
	if err != nil {
		return zero, err
	}

	vu, err := time.Parse("2006-01-02T15:04:05", c.ValidUntil)
	if err != nil {
		return zero, err
	}

	if !vu.After(t) {
		return zero, fmt.Errorf("valid_until must be after datetime")
	}

	addr := address.Address{Number: "0"}
	if c.Address != nil {
		addr, err = c.Address.Prepare()
		if err != nil {
			return zero, err
		}
	}

	return PickupLocationBlueprint{
		Name:        c.Name,
		Description: c.Description,
		Datetime:    t,
		ValidUntil:   vu,
		Address:     addr,
		ShopID:      shopID,
	}, nil
}

type UpdatePickupLocationRequest struct {
	Name        *string                          `json:"name"`
	Description *string                          `json:"description"`
	DateTime    *string                          `json:"datetime"`
	ValidUntil  *string                          `json:"validUntil"`
	Address     *address.AddressRequestOptional `json:"address"`
}
