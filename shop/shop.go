package shop

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/authmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func ViewHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		shopID := c.Param("shop")

		shop, err := shopmodel.Find(
			c.Request.Context(),
			pool, shopID,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.Handle404(c)
				return
			}

			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"shop": shop.ToResource(),
		})
	}
}

func ListHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		shops, err := shopmodel.Get(c.Request.Context(), pool)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.Handle404(c)
				return
			}

			util.Handle500(c, err)
			return
		}

		resources := make([]shopmodel.ShopResource, 0, len(shops))

		for _, shop := range shops {
			resources = append(resources, shop.ToResource())
		}

		c.JSON(200, gin.H{
			"shops": resources,
		})
	}
}

func CreateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var (
			shop shopmodel.CreateShopRequest
			user int
		)

		err := c.ShouldBindJSON(&shop)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		user, err = authmodel.GetUserFromContext(c)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		shopID, err := shopmodel.Create(
			c.Request.Context(), pool,
			user,
			shop,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"message": "estabelecimento criado com sucesso",
			"shop": shopmodel.ShopSummaryResource{
				ID:        shopID,
				Name:      shop.Name,
				Slug:      slug.Make(shop.Name),
				Telephone: shop.Telephone,
			},
		})
	}
}
