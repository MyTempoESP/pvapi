package saque

import (
	"context"
	"fmt"
	"strconv"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/financemodel"
	"linhazero.com/pvapi/models/saquemodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

var (
	ErrInsufficientFunds = errors.New("saldo indisponivel para saque")
)

func createSaqueGuarded(
	ctx context.Context,
	tx db.Queryable,
	user authmodel.AuthUser,
	request saquemodel.CreateSaqueRequest,
) (int, error) {
	available, err := financemodel.AvailableBalance(
		ctx, tx,
		user.ActiveRoleID,
	)

	if err != nil {
		return 0, err
	}

	if request.AskedValue.GreaterThan(available) {
		return 0, ErrInsufficientFunds
	}

	id, err := saquemodel.Create(
		ctx, tx,
		user.ActiveRoleID,
		user.UserID,
		request,
	)

	return id, err
}

func CreateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request saquemodel.CreateSaqueRequest

		user := authmodel.GetUserFromContext(c)

		err := c.ShouldBindJSON(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		if request.AskedValue.LessThanOrEqual(decimal.NewFromInt(0)) {
			util.HandleBadRequest(c, fmt.Errorf("valor do saque deve ser positivo"))
			return
		}

		ok, err := shopmodel.HasBankAccount(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
			request.BankAccountID,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		if !ok {
			util.Handle404(c, nil)
			return
		}

		id, err := util.TransactionReturn(
			c.Request.Context(), pool,
			func (
				ctx context.Context,
				tx db.Queryable,
			) (int, error) {
				return createSaqueGuarded(
					ctx, tx,
					user, request,
				)
			},
		)

		if err != nil {
			if errors.Is(err, ErrInsufficientFunds) {
				util.HandleBadRequest(c, err)
				return
			}

			util.Handle500(c, err)
			return
		}

		util.WrapLabeled(c, "saque solicitado com sucesso",
			"saque", gin.H{
				"id":     id,
				"status": "pending",
			},
		)
	}
}

func ListHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		rows, err := saquemodel.ListByShop(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		saques := util.ResourceCollection(rows)

		util.WrapLabeled(c, "saques listados com sucesso", "saques", saques)
	}
}

func ViewHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("saque"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		saque, err := saquemodel.FindByID(
			c.Request.Context(),
			pool,
			id,
		)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(c, "saque encontrado com sucesso", "saque", saque.ToResource())
	}
}

func UpdateStatusHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req saquemodel.UpdateSaqueRequest

		id, err := strconv.Atoi(c.Param("saque"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = c.ShouldBindJSON(&req)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = saquemodel.UpdateStatus(
			c.Request.Context(),
			pool,
			id,
			req.Status,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.OK(c, "status do saque atualizado com sucesso")
	}
}

func DeleteHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("saque"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = saquemodel.Delete(
			c.Request.Context(),
			pool,
			id,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.OK(c, "saque removido com sucesso")
	}
}
