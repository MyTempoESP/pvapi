package pickupmodel

import "linhazero.com/pvapi/address"

type PickupLocationResource struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	DateTime string          `json:"datetime"`
	Address  address.Address `json:"address"`
}

func (p PickupLocation) ToResource() PickupLocationResource {
	return PickupLocationResource{
		ID:       p.ID,
		Name:     p.Name,
		DateTime: p.Datetime.Format("2006-01-02T15:04:05"),
		Address:  p.Address,
	}
}
