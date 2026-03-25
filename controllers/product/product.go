package product

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/shoputils"
)

func ListByShopHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	type Pcollect = []productmodel.Product

	return func(c *gin.Context) {
		products, err := shoputils.WithShopFromContextReturn(
			c, pool,
			func(ctx context.Context, shop *shopmodel.Shop) (Pcollect, error) {
				return productmodel.GetByShop(
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
		var (
			product productmodel.CreateProductRequest
			userID  int
			err     error
		)

		userID, err = authmodel.GetUserFromContext(c)

		if err != nil {
			return
		}

		err = c.ShouldBind(&product)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		blueprint, err := product.Prepare(c)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		shop, err := shopmodel.FindByUser(
			c.Request.Context(),
			pool, userID,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.Handle404(c)
				return
			}

			util.Handle500(c, err)
			return
		}

		_, err = productmodel.Create(
			c.Request.Context(), pool,
			blueprint, shop.ID,
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
