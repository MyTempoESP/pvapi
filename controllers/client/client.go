package client

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/clientmodel"
	"linhazero.com/pvapi/util"
)

func ListHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser := authmodel.GetUserFromContext(c)

		shopID := authUser.ActiveRoleID

		clients, err := clientmodel.List(c.Request.Context(), pool, shopID)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.WrapLabeled(
			c, "clientes listados com sucesso",
			"clients", util.ResourceCollection(clients),
		)
	}
}
