package pickupmodel

import "linhazero.com/pvapi/address"

type CreatePickupLocationRequest struct {
	Name     string `json:"name"`
	DateTime string `json:"datetime"`
	Address  *address.Address
}
