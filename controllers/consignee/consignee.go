package consignee

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/consigneemodel"
	"linhazero.com/pvapi/util"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		consignees, err := consigneemodel.List(
			c.Request.Context(), pool, user.ActiveRoleID)

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

		user := authmodel.GetUserFromContext(c)

		consignee, err := consigneeRequest.Prepare(user.ActiveRoleID)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		id, err := util.TransactionReturn[int](
			c.Request.Context(),
			pool,
			consignee.Save,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"message": "monitor criado com sucesso",
			"monitor": gin.H{
				"id": id,
				"name": consignee.Name,
				"email": consignee.Email,
				"phone": consignee.Phone,
				"address": consignee.Address,
			},
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
