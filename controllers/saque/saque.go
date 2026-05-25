package saque

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func CreateHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req shopmodel.CreateSaqueRequest

		user := authmodel.GetUserFromContext(c)

		err := c.ShouldBindJSON(&req)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		id, err := shopmodel.CreateSaque(
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

		saques, err := shopmodel.ListSaquesByShop(
			c.Request.Context(),
			pool,
			user.ActiveRoleID,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.WrapLabeled(c, "saques listados com sucesso", "saques", saques)
	}
}
