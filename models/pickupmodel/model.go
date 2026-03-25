package pickupmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/address"
)

type PickupLocation struct {
	Name     string `json:"name"`
	Datetime string `json:"datetime"`
	Address  *address.Address
}

func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	pickupLocation CreatePickupLocationRequest,
) (PickupLocation, error) {
	return PickupLocation{}, nil
}
