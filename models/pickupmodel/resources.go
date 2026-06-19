package pickupmodel

import "linhazero.com/pvapi/address"

type PickupLocationResource struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	DateTime    string           `json:"datetime"`
	ValidUntil  string           `json:"validUntil"`
	Address     *address.Address `json:"address,omitempty"`
}

func (p PickupLocation) ToResource() PickupLocationResource {
	return PickupLocationResource{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		DateTime:    p.Datetime.Format("2006-01-02T15:04:05"),
		ValidUntil:  p.ValidUntil.Format("2006-01-02T15:04:05"),
		Address:     p.Address.ToResource(),
	}
}
