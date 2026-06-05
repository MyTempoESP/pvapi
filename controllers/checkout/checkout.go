package checkout

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/checkoutmodel"
	"linhazero.com/pvapi/models/clientmodel"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/checkoutclient"
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

		// Resolve identity from raw request fields before Confirm
		ci := checkoutclient.Resolve(
			req.ClientEmail,
			req.Telephone,
			req.CPF,
			req.ClientName,
		)

		// Effective email (real or synthetic) participates in content
		// hash for dedup, preventing cross-customer collisions.
		req.ClientEmail = ci.EffectiveEmail

		resp, err := checkoutmodel.Confirm(
			c.Request.Context(),
			pool,
			req,
		)

		if err != nil {
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

		// Best-effort client association after response is sent
		if ci.EffectiveEmail != "" {
			associateClientAfterCheckout(c, pool, ci, resp)
		}
	}
}

func associateClientAfterCheckout(
	c *gin.Context,
	pool *pgxpool.Pool,
	ci checkoutclient.CheckoutClientOutput,
	resp checkoutmodel.ConfirmResponse,
) {
	hash, err := util.HashPassword(ci.Password)
	if err != nil {
		util.LogError(c, "failed to hash password", "error", err)
		return
	}

	ctx := context.Background()

	isNew, err := clientmodel.CreateAndAssociateByCheckout(
		ctx, pool,
		ci.ClientName, ci.EffectiveEmail,
		hash,
		resp.CheckoutID,
		ci.NormalizedPhone,
		ci.ValidatedCPF,
	)
	if err != nil {
		util.LogError(c, "failed to create client from checkout",
			"error", err,
			"checkout_id", resp.CheckoutID,
		)
		return
	}

	util.LogInfo(c, "client created/reused from checkout",
		"checkout_id", resp.CheckoutID,
		"is_new", isNew,
	)
}
