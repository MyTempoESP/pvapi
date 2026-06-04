package product

import (
	"errors"
	"context"
	"os"
	"fmt"
	"strconv"

//~extern
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/util"
)

func PublicListingHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		rawShopIDs := c.Request.URL.Query()["shop"]
		rawConsignmentIDs := c.Request.URL.Query()["consignment"]

		shopIDs := parseQueryInts(rawShopIDs)
		consignmentIDs := parseQueryInts(rawConsignmentIDs)

		products, err := productmodel.PublicIndex(
			c.Request.Context(), pool,
			shopIDs, consignmentIDs,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		// Batch-fetch pickups for all listed products
		productIDs := make([]int, len(products))
		for i, p := range products {
			productIDs[i] = p.ID
		}

		pickupMap, err := productmodel.BatchFetchPickups(
			c.Request.Context(), pool, productIDs,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		for i := range products {
			products[i].Pickups = pickupMap[products[i].ID]
		}

		util.WrapMap(
			c, "Produtos listados com sucesso!",
			gin.H{
				"produtos": util.ResourceCollection(products),
			},
		)
	}
}

func parseQueryInts(raw []string) []int {
	if len(raw) == 0 {
		return nil
	}
	ints := make([]int, 0, len(raw))
	for _, s := range raw {
		n, err := strconv.Atoi(s)
		if err == nil {
			ints = append(ints, n)
		}
	}
	return ints
}

func DeleteHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		product, err := strconv.Atoi(c.Params.ByName("product"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = util.TransactionAware(
			c.Request.Context(), pool,
			func (ctx context.Context, tx pgx.Tx) error {
				return productmodel.Delete(ctx, tx, product)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "Produto deletado com sucesso")
	}
}

func DeleteImageHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		product, err := strconv.Atoi(c.Params.ByName("product"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		imageBase := c.Params.ByName("url")

		if imageBase == "" {
			util.HandleBadRequest(
				c, errors.New("nenhuma imagem especificada"))
			return
		}

		imageUrl, err := util.GetImageURL(imageBase)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		dest := fmt.Sprintf("./public/uploads/%s", imageBase)

		hasImage, err := productmodel.HasImage(
			c.Request.Context(), pool,
			product, imageUrl,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		if !hasImage {
			util.HandleBadRequest(
				c, errors.New("o produto nao possui a imagem"))
			return
		}

		util.LogInfo(c, "Removing an image file", "path", dest)
		err = os.Remove(dest)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.LogInfo(c, "Removing image from product")
		err = util.TransactionAware(
			c.Request.Context(), pool,
			func (ctx context.Context, tx pgx.Tx) error {
				return productmodel.DeleteImage(
					ctx, tx,
					product, imageUrl,
				)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "Imagem deletada com sucesso")
	}
}

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
	newStatus string,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request productmodel.ModerateProductRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		productID, err := strconv.Atoi(c.Param("product"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = productmodel.Moderate(
			c.Request.Context(), pool,
			request,
			productID,
			newStatus,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "Product "+newStatus)
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
		productID, err := strconv.Atoi(c.Param("product"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		product, err := productmodel.PublicFind(
			c.Request.Context(), pool,
			productID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		// Fetch pickups for this product
		pickupMap, err := productmodel.BatchFetchPickups(
			c.Request.Context(), pool, []int{productID},
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		product.Pickups = pickupMap[productID]

		util.Wrap(
			c, "produto encontrado!",
			product.ToResource(),
		)
	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var product productmodel.UpdateProductRequest

		user := authmodel.GetUserFromContext(c)

		productID, err := strconv.Atoi(c.Param("product"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = c.ShouldBind(&product)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		blueprint, err := product.Prepare(
			c,
			productID,
			user.ActiveRoleID,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		err = util.TransactionAware(c, pool, blueprint.Exec)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		updatedProduct, err := productmodel.Find(
			c.Request.Context(), pool, productID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.Wrap(
			c, "Produto editado com sucesso!",
			updatedProduct.ToResource(),
		)
	}
}
