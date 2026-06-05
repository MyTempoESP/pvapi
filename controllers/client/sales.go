package client

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/clientmodel"
	"linhazero.com/pvapi/util"
)

// SalesHandler returns sales for a given client, scoped to the authenticated
// shop. Only shop users can access this endpoint.
func SalesHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser := authmodel.GetUserFromContext(c)

		if authUser.Role != "shop" {
			util.Handle404(c, fmt.Errorf("not a shop user"))
			return
		}

		clientID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			util.Handle404(c, fmt.Errorf("invalid client id"))
			return
		}

		exists, err := clientmodel.ClientExists(
			c.Request.Context(), pool, clientID,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}
		if !exists {
			util.Handle404(c, fmt.Errorf("client not found"))
			return
		}

		shopID := authUser.ActiveRoleID

		sales, err := clientmodel.ListClientSales(
			c.Request.Context(),
			pool,
			clientID,
			shopID,
		)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		// Return bare top-level JSON array
		if sales == nil {
			c.JSON(http.StatusOK, []any{})
			return
		}
		c.JSON(http.StatusOK, sales)
	}
}
