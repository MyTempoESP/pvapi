package shop

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func ViewHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		shopID := c.Param("shop")

		shop, err := shopmodel.FindDetailed(
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

		resources := util.ResourceCollection(shops)

		c.JSON(200, gin.H{
			"shops": resources,
		})
	}
}

func CreateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var (
			shop shopmodel.CreateShopRequest
		)

		err := c.ShouldBindJSON(&shop)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		shopID, err := shopmodel.Create(
			c.Request.Context(), pool, shop,
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

func UpdateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var (
			shop shopmodel.UpdateShopRequest
		)

		err := c.ShouldBindJSON(&shop)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		shopIDString := c.Param("shop")
		shopID, err := strconv.Atoi(shopIDString)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		err = shopmodel.Update(
			c.Request.Context(), pool,
			shopID,
			shop,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"message": "estabelecimento editado com sucesso",
		})
	}
}
