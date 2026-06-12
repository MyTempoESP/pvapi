package pickups

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/pickupmodel"
	"linhazero.com/pvapi/util"
)

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request pickupmodel.CreatePickupLocationRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		user := authmodel.GetUserFromContext(c)
		pickup, err := request.Prepare(user.ActiveRoleID)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		id, err := util.TransactionReturn(
			c, pool,
			pickup.Save,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.WrapMap(
			c, "ponto criado com sucesso",
			gin.H{
				"ponto": id,
				"id": id,
			},
		)
	}
}

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		pickups, err := pickupmodel.List(
			c.Request.Context(), pool, user.ActiveRoleID)

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

func DeleteHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("pickup"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = pickupmodel.Delete(
			c.Request.Context(),
			pool,
			id,
		)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "ponto de entrega deletado com sucesso")
	}
}

func UpdateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("pickup"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		var req pickupmodel.UpdatePickupLocationRequest

		err = c.ShouldBindJSON(&req)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = pickupmodel.Update(
			c.Request.Context(),
			pool,
			id,
			req,
		)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "ponto de entrega atualizado com sucesso")
	}
}
