package product

import (
	"errors"
	"context"
	"os"
	"fmt"
	"strconv"

//~extern
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/consignmentmodel"
	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/util"
)

func PublicListingHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		token := c.Request.URL.Query().Get("token")

		if token == "" {
			util.WrapMap(
				c, "Produtos listados com sucesso!",
				gin.H{
					"produtos":  []productmodel.ProductConsignmentResource{},
					"retiradas": []productmodel.ConsignmentPickupResource{},
				},
			)
			return
		}

		// Validate UUID format
		if _, err := uuid.FromString(token); err != nil {
			util.HandleBadRequest(c, errors.New("token inválido: formato UUID esperado"))
			return
		}

		// Find consignment by code
		lookup, err := consignmentmodel.FindByCode(c.Request.Context(), pool, token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.Handle404(c, fmt.Errorf("remessa com código %q não encontrada", token))
				return
			}
			util.Handle500(c, err)
			return
		}

		if lookup.Status == "ARQUIVADO" {
			util.HandleGone(c, "remessa arquivada, não está mais disponível")
			return
		}

		if lookup.Status == "PAUSADO" {
			util.Handle404(c, fmt.Errorf("remessa temporariamente indisponível"))
			return
		}

		// Fetch products
		products, err := productmodel.PublicIndexByConsignmentCode(
			c.Request.Context(), pool, token,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		// Fetch pickups
		pickups, err := productmodel.PickupsByConsignmentCode(
			c.Request.Context(), pool, token,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.WrapMap(
			c, "Produtos listados com sucesso!",
			gin.H{
				"produtos": util.ToResourceCollection(
					products,
					productmodel.ConsignmentProductRow.ToConsignmentResource,
				),
				"retiradas": util.ToResourceCollection(
					pickups,
					productmodel.PickupByCodeRow.ToConsignmentPickupResource,
				),
			},
		)
	}
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
