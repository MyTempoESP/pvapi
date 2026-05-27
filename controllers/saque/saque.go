package saque

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/saquemodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func CreateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req saquemodel.CreateSaqueRequest

		user := authmodel.GetUserFromContext(c)

		err := c.ShouldBindJSON(&req)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		ok, err := shopmodel.HasBankAccount(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
			req.BankAccountID,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}
		if !ok {
			util.Handle404(c, nil)
			return
		}

		id, err := saquemodel.Create(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
			user.UserID,
			req,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.WrapLabeled(c, "saque solicitado com sucesso", "saque", gin.H{
			"id":     id,
			"status": "pending",
		})
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
