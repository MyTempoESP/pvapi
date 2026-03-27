package pickupmodel

import (
	"time"

	"linhazero.com/pvapi/address"
)

type PickupLocationBlueprint struct {
	ShopID   int
	Name     string
	Datetime time.Time
	Address  address.Address
}
