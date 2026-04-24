package cart

import (
	"strconv"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/models/cartmodel"
	"linhazero.com/pvapi/util"
)

func RemoveHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		item, err := strconv.Atoi(c.Params.ByName("item"))

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = util.TransactionAware(
			c.Request.Context(), pool,
			func (ctx context.Context, tx pgx.Tx) error {
				return cartmodel.RemoveCartProduct(
					ctx, tx,
					item,
				)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "produto removido do carrinho")
	}
}

func ViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionID, ok := c.GetQuery("session_id")
		if !ok {
			util.HandleBadRequest(c, fmt.Errorf("No session_id"))
			return
		}

		cart, err := cartmodel.FindBySession(
			c.Request.Context(), pool,
			sessionID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		resource := cart.ToResource(sessionID)

		util.WrapMap(c, "Dados do carrinho obtidos com sucesso!",
			gin.H{
				"items": resource.Items,
				"total": resource.TotalF, // XXX: using floats temporarily for compat
				"item_count": resource.Count,
				"total_with_discount": resource.TotalWithDiscountF, // XXX: __
				"total_discount": resource.TotalDiscountF, // XXX: __
				"session_id": resource.SessionID,
			},
		)
	}
}

func AddProductHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request cartmodel.AddToCartRequest

		err := c.ShouldBind(&request)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = util.TransactionAware(
			c.Request.Context(), pool,
			request.Exec,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "produto adicionado!")
	}
}
