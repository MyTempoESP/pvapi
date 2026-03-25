package pickups

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/pickupmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/shoputils"
)

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	type PickupLocation = pickupmodel.PickupLocation

	return func(c *gin.Context) {
		var request pickupmodel.CreatePickupLocationRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		pickup, err := shoputils.WithShopFromContextReturn[PickupLocation](
			c, pool,
			func(ctx context.Context, s *shopmodel.Shop) (PickupLocation, error) {
				return pickupmodel.Create(ctx, pool, request)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c, "ponto criado com sucesso",
			"ponto", pickup,
		)
	}
}

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		util.WrapLabeled(
			c,
			"pontos de entrega listados com sucesso",
			"data", []any{},
		)
	}
}
