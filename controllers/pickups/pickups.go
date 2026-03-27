package pickups

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/pickupmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/pickuputils"
	"linhazero.com/pvapi/util/shoputils"
)

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		pickup, err := pickuputils.NewPickupHandler(
			c,
			pool,
		)

		if err != nil {
			util.HandleGuess(c, err)
			return
		}

		pickupID, err := shoputils.WithShopFromContextReturn(
			c, pool,
			pickup.Save,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c, "ponto criado com sucesso",
			"ponto", pickupmodel.PickupLocationResource{
				ID:       pickupID,
				Name:     pickup.Name,
				DateTime: pickup.DateTime,
				Address:  *pickup.Address,
			},
		)
	}
}

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	type Pickups = []pickupmodel.PickupLocation

	return func(c *gin.Context) {
		pickups, err := shoputils.WithShopFromContextReturn(
			c, pool,
			func(ctx context.Context, s *shopmodel.Shop) (Pickups, error) {
				return pickupmodel.List(ctx, pool, s.ID)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		collection := util.ResourceCollection(pickups)

		util.Wrap(
			c,
			"pontos de entrega listados com sucesso",
			collection,
		)
	}
}
