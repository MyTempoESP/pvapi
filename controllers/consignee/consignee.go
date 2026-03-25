package consignee

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/consigneemodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/shoputils"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	type Consignees = []consigneemodel.Consignee

	return func(c *gin.Context) {
		consignees, err := shoputils.WithShopFromContextReturn(
			c, pool,
			func(ctx context.Context, shop *shopmodel.Shop) (Consignees, error) {
				return consigneemodel.List(
					ctx,
					pool,
					shop.ID,
				)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		collection := util.ResourceCollection(consignees)

		util.Wrap(
			c, "monitores listados com sucesso",
			collection,
		)
	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		// TODO: generalize
		var consigneeRequest consigneemodel.CreateConsigneeRequest

		err := c.ShouldBind(&consigneeRequest)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		shop, err := shoputils.FindFromContext(c, pool)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		consignee, err := consigneeRequest.Prepare(shop.ID)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		err = util.Transaction(
			c, pool,
			consignee.Save,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"message": "monitor criado com sucesso",
		})
	}
}

func ViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}
