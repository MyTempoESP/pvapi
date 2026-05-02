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

func ClearHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request struct{
			SessionID string `json:"session_id"`
		}

		err := c.ShouldBind(&request);

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = util.TransactionAware(
			c.Request.Context(), pool,
			func (ctx context.Context, tx pgx.Tx) error {
				return cartmodel.ClearSession(
					ctx, tx, request.SessionID)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "carrinho limpo")
	}
}

func getItem(c *gin.Context) (int, error) {
	return strconv.Atoi(c.Params.ByName("item"))
}

func QuantityHandler(
	pool *pgxpool.Pool,
) func (c *gin.Context) {
	return func (c *gin.Context) {
		item, err := getItem(c)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		var request struct{
			Quantity int `json:"quantity" binding:"required"`

			//TODO: validate session
			//SessionID string `json:"session_id" binding:"required"`
		}

		err = c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = util.TransactionAware(
			c.Request.Context(), pool,
			func (ctx context.Context, tx pgx.Tx) error {
				return cartmodel.ItemNewQuantity(
					ctx, tx,
					item,
					request.Quantity,
					//request.SessionID,
				)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "quantidade alterada com sucesso")
	}
}

func RemoveHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		item, err := getItem(c)

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
				"total": resource.TotalF,
				"item_count": resource.Count,
				"total_with_discount": resource.TotalWithDiscountF,
				"total_discount": resource.TotalDiscountF,
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
