package checkout

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/checkoutmodel"
	"linhazero.com/pvapi/util"
)

func ConfirmHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var req checkoutmodel.ConfirmRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		resp, err := checkoutmodel.Confirm(
			c.Request.Context(),
			pool,
			req,
		)

		if err != nil {
			switch {
			case errors.Is(err, checkoutmodel.ErrCartNotFound):
				util.Handle404(c, err)
			case errors.Is(err, checkoutmodel.ErrCartEmpty),
				errors.Is(err, checkoutmodel.ErrProductNotInConsignment):
				c.AbortWithStatusJSON(422, gin.H{
					"success": false,
					"message": err.Error(),
				})
			case errors.Is(err, checkoutmodel.ErrInsufficientStock):
				c.AbortWithStatusJSON(409, gin.H{
					"success": false,
					"message": err.Error(),
				})
			default:
				util.Handle500(c, err)
			}
			return
		}

		c.JSON(200, gin.H{
			"success":   true,
			"message":   "checkout confirmed",
			"unique_id": resp.UniqueID,
			"status":    resp.Status,
		})
	}
}
