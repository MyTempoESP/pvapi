package product

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func ListByShopHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, err := authmodel.GetUserFromContext(c)

		ctx := c.Request.Context()

		if err != nil {
			util.Handle500(c, err)
			return
		}

		shop, err := shopmodel.FindByUser(ctx, pool, user)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.Handle404(c)
				return
			}

			util.Handle500(c, err)
			return
		}

		products, err := productmodel.GetByShop(ctx, pool, shop.ID)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		resources := util.ResourceCollection(products)

		c.JSON(200, gin.H{
			"products": resources,
		})
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
		var (
			product productmodel.CreateProductRequest
		)

		err := c.ShouldBindJSON(&product)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		productID, err := productmodel.Create(
			c.Request.Context(), pool,
			product,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{})
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
