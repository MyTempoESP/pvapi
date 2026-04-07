package product

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/util"
)

func ListPendingModerationHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		products, err := productmodel.PendingIndex(
			c.Request.Context(), pool,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		collection := util.ToResourceCollection(
			products,
			productmodel.Product.ToPendingResource,
		)

		util.Wrap(
			c, "produtos pendentes listados com sucesso",
			collection,
		)
	}
}

func ModerationHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
	}
}

func ListByShopHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		products, err := productmodel.GetByShop(
			c.Request.Context(), pool,
			user.ActiveRoleID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		collection := util.ResourceCollection(products)

		util.Wrap(
			c, "produtos listados com sucesso",
			collection,
		)
	}
}

func ListByParamHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var product productmodel.CreateProductRequest

		user := authmodel.GetUserFromContext(c)

		err := c.ShouldBind(&product)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		blueprint, err := product.Prepare(c, user.ActiveRoleID)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		err = util.TransactionAware(c, pool, blueprint.Save)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Produto(s) criado(s) com sucesso!",
		})
	}
}

func PublicViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {

	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		// TODO
	}
}
