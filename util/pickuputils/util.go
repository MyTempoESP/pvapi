package pickuputils

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/pickupmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

type newPickup struct {
	Save func(context.Context, *shopmodel.Shop) (int, error)

	pickupmodel.CreatePickupLocationRequest
}

func NewPickup(
	c *gin.Context,
	pool *pgxpool.Pool,
	request pickupmodel.CreatePickupLocationRequest,
) (newPickup, error) {
	return newPickup{
		Save: func(
			ctx context.Context,
			s *shopmodel.Shop,
		) (id int, err error) {
			pickup, err := request.Prepare(s.ID)

			if err != nil {
				return
			}

			util.Log(
				c,
				"name", pickup.Name,
				"shop_id", pickup.ShopID,
				"datetime", pickup.Datetime,
			)

			return util.TransactionReturn(ctx, pool, pickup.Save)
		},
	}, nil
}

func NewPickupHandler(
	c *gin.Context,
	pool *pgxpool.Pool,
) (newPickup, error) {
	var (
		zero    newPickup
		request pickupmodel.CreatePickupLocationRequest
	)

	err := c.ShouldBind(&request)

	if err != nil {
		return zero, fmt.Errorf(
			"malformed pickup creation request: %w", util.BadRequestError)
	}

	return NewPickup(c, pool, request)
}
