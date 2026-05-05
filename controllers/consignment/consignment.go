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
		user := authmodel.GetUserFromContext(c)

		consignments, err := consignmentmodel.ListAll(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
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

func ListProductsHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		consignmentID := c.Param("consignment")

		products, err := consignmentmodel.ListProducts(
			c.Request.Context(),
			pool,
			consignmentID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		collection := util.ResourceCollection(products)

		util.WrapMap(
			c, "produtos listados com sucesso",
			gin.H{
				"products":       collection,
				"total_products": len(collection),
				"consignment_id": consignmentID,
			},
		)
	}
}

func AddProductHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request consignmentmodel.AddProductRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		consignmentID, err := strconv.Atoi(c.Param("consignment"))

		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.LogInfo(
			c,
			"Trying to add products to consignment",
			"Consignment", consignmentID,
		)

		action := request.Prepare(consignmentID)

		err = util.TransactionAware(
			c.Request.Context(), pool,
			action.Exec,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "Produto adicionado na remessa com sucesso")
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

		consignment, err := request.Prepare(
			user.ActiveRoleID, // ShopID
		)

		if err != nil {
			util.HandleGuess(c, err)
			return
		}

		id, err := util.TransactionReturn(
			c.Request.Context(), pool,
			consignment.Save,
		)

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

		consignment, err := consignmentmodel.FindAny(
			c.Request.Context(),
			pool,
			consignmentID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.Wrap(
			c, "remessa acessada com sucesso",
			consignment.ToResource(),
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

		util.OK(c, "remessa atualizada com sucesso")
	}
}

func UpdateProductHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request consignmentmodel.UpdateConsignmentProductRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		consignment, err := strconv.Atoi(c.Param("consignment"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		product, err := strconv.Atoi(c.Param("product"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		prepare := request.Prepare(consignment, product)

		err = util.TransactionAware(
			c.Request.Context(), pool,
			prepare.Exec,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "Registro na remessa editado com sucesso!")
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
