package consignment

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/consignmentmodel"
	"linhazero.com/pvapi/util"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		consignments, err := consignmentmodel.List(
			c.Request.Context(),
			pool,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		collection := util.ResourceCollection(consignments)

		util.Wrap(
			c, "remessas listados com sucesso",
			collection,
		)
	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request consignmentmodel.CreateConsignmentRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		user := authmodel.GetUserFromContext(c)

		consignment, err := request.Prepare(user.ActiveRoleID)

		if err != nil {
			util.HandleGuess(c, err)
			return
		}

		id, err := consignment.Save(c.Request.Context(), pool)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c, "remessa criada com sucesso",
			"id", id,
		)
	}
}

func ViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		consignmentID := c.Param("consignment")

		consignment, err := consignmentmodel.Find(
			c.Request.Context(),
			pool,
			consignmentID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c, "remessa acessada com sucesso",
			"consignment", consignment.ToResource(),
		)
	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		consignmentID := c.Param("consignment")

		var request consignmentmodel.UpdateConsignmentRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		id, err := strconv.Atoi(consignmentID)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		consignment := request.Prepare(id)

		err = consignment.Save(c.Request.Context(), pool)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c, "remessa atualizada com sucesso",
			"consignment", consignment.ToResource(),
		)
	}
}

func DeleteHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := consignmentmodel.Delete(
			c.Request.Context(), pool,
			c.Param("consignment"),
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.OK(
			c, "remessa deletada com sucesso",
		)
	}
}
