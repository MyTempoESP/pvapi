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

		util.Wrap(
			c, "clientes listados com sucesso",
			util.ResourceCollection(clients),
		)
	}
}
