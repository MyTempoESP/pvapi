package bankaccount

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/bankaccountmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func CreateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req bankaccountmodel.CreateBankAccountRequest

		user := authmodel.GetUserFromContext(c)

		err := c.ShouldBindJSON(&req)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		id, err := bankaccountmodel.Create(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
			req,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.WrapLabeled(c, "conta bancária criada com sucesso", "conta", gin.H{
			"id":     id,
			"status": "active",
		})
	}
}

func ListHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		rows, err := bankaccountmodel.ListByShop(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		contas := util.ResourceCollection(rows)

		util.WrapLabeled(c, "contas bancárias listadas com sucesso", "contas", contas)
	}
}

func UpdateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req bankaccountmodel.UpdateBankAccountRequest

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		user := authmodel.GetUserFromContext(c)

		ok, err := shopmodel.HasBankAccount(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
			id,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}
		if !ok {
			util.Handle404(c, nil)
			return
		}

		err = c.ShouldBindJSON(&req)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		err = bankaccountmodel.Update(
			c.Request.Context(),
			pool,
			id,
			req,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.OK(c, "conta bancária atualizada com sucesso")
	}
}

func DeleteHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		user := authmodel.GetUserFromContext(c)

		ok, err := shopmodel.HasBankAccount(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
			id,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}
		if !ok {
			util.Handle404(c, nil)
			return
		}

		err = bankaccountmodel.Delete(
			c.Request.Context(),
			pool,
			id,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.OK(c, "conta bancária removida com sucesso")
	}
}
