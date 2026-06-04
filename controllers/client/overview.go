package client

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/clientmodel"
	"linhazero.com/pvapi/util"
)

func OverviewHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser := authmodel.GetUserFromContext(c)

		if authUser.Role != "client" {
			util.Handle404(c, fmt.Errorf("user is not a client"))
			return
		}

		clientID := authUser.ActiveRoleID

		overview, err := clientmodel.GetOverview(
			c.Request.Context(),
			pool,
			clientID,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"success":  true,
			"message":  "visão geral carregada",
			"orders":   overview.Orders,
			"pending":  overview.Pending,
			"packages": overview.Packages,
		})
	}
}
