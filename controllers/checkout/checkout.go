package checkout

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/middleware/apitoken"
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

		// Require at least one identity field (client_email, cpf, or telephone)
		if strings.TrimSpace(req.ClientEmail) == "" &&
			strings.TrimSpace(req.CPF) == "" &&
			strings.TrimSpace(req.Telephone) == "" {
			c.AbortWithStatusJSON(400, gin.H{
				"success": false,
				"message": "Informe email, CPF ou telefone para continuar",
			})
			return
		}

		apiClientID := apitoken.GetAPIClientID(c)

		resp, err := checkoutmodel.Confirm(
			c.Request.Context(),
			pool,
			req,
			apiClientID,
		)

		if err != nil {
			var rlInfo *checkoutmodel.RateLimitInfo
			if errors.As(err, &rlInfo) {
				c.AbortWithStatusJSON(429, gin.H{
					"success":      false,
					"message":      rlInfo.Message,
					"retry_after":  rlInfo.RetryAfter,
				})
				return
			}

			switch {
			case errors.Is(err, checkoutmodel.ErrPackageNotFound):
				c.AbortWithStatusJSON(422, gin.H{
					"success": false,
					"message": err.Error(),
				})
			case errors.Is(err, checkoutmodel.ErrInsufficientStock):
				c.AbortWithStatusJSON(409, gin.H{
					"success": false,
					"message": err.Error(),
				})
			case errors.Is(err, checkoutmodel.ErrNoIdentity):
				c.AbortWithStatusJSON(400, gin.H{
					"success": false,
					"message": "Informe email, CPF ou telefone para continuar",
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
			"qr_codes":  resp.QrCodes,
		})
	}
}
