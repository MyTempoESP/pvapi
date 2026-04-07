package shop

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func PublicListProductsHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		slug := c.Param("shop")

		shopID, err := shopmodel.Resolve(
			c.Request.Context(), pool,
			slug,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		products, err := productmodel.PublicIndexByShop(
			c.Request.Context(), pool,
			shopID,
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

func ViewHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		shopIdentification := c.Param("shop")

		shopID, err := strconv.Atoi(shopIdentification)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		shop, err := shopmodel.FindDetailed(
			c.Request.Context(),
			pool, shopID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c,
			"lojas listadas com sucesso",
			"shop", shop.ToResource(),
		)
	}
}

func ListHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		shops, err := shopmodel.List(c.Request.Context(), pool)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		collection := util.ResourceCollection(shops)

		// No 'data' wrapper label
		util.WrapLabeled(
			c,
			"lojas listadas com sucesso",
			"shops", collection,
		)
	}
}

func CreateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	// TODO TEST: verify the success status works correctly for creating

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

		util.WrapLabeled(
			c,
			"estabelecimento criado com sucesso",
			"shop", shopmodel.SummarizeRequest(shopID, shop),
		)
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
