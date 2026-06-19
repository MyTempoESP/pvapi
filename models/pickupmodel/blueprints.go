package pickupmodel

import (
	"time"

	"linhazero.com/pvapi/address"
)

type PickupLocationBlueprint struct {
	ShopID      int
	Name        string
	Description string
	Datetime    time.Time
	ValidUntil  time.Time
	Address     address.Address
}
