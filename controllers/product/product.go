package product

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/util"
)

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

		blueprint, err := product.Prepare(c)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		_, err = productmodel.Create(
			c.Request.Context(),
			pool,
			blueprint,
			user.ActiveRoleID,
		)

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
		// TODO
	}
}
