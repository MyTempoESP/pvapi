package checkout

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/env"
	"linhazero.com/pvapi/mail"
	"linhazero.com/pvapi/models/checkoutmodel"
	"linhazero.com/pvapi/models/clientmodel"
	"linhazero.com/pvapi/util"
)

func ConfirmHandler(
	pool *pgxpool.Pool,
	m mail.Mailer,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		sessionID := c.Query("session_id")
		if sessionID == "" {
			util.HandleBadRequest(c, errors.New("session_id is required"))
			return
		}

		var req checkoutmodel.ConfirmRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		resp, err := checkoutmodel.Confirm(
			c.Request.Context(),
			pool,
			sessionID,
			req,
		)

		if err != nil {
			switch {
			case errors.Is(err, checkoutmodel.ErrCartNotFound):
				util.Handle404(c, err)
			case errors.Is(err, checkoutmodel.ErrCartEmpty),
				errors.Is(err, checkoutmodel.ErrPackageNotFound):
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
			"qr_code":   resp.QrCode,
		})

		// Best-effort client creation after response is sent
		if req.ClientEmail != "" {
			createClientAfterCheckout(c, pool, m, req, resp)
		}
	}
}

func createClientAfterCheckout(
	c *gin.Context,
	pool *pgxpool.Pool,
	m mail.Mailer,
	req checkoutmodel.ConfirmRequest,
	resp checkoutmodel.ConfirmResponse,
) {
	clientName := req.ClientName
	if clientName == "" {
		clientName = util.DeriveClientName(req.ClientEmail)
	}

	plaintext, err := util.GenerateRandomPassword(16)
	if err != nil {
		util.LogError(c, "failed to generate random password", "error", err)
		return
	}

	hash, err := util.HashPassword(plaintext)
	if err != nil {
		util.LogError(c, "failed to hash password", "error", err)
		return
	}

	ctx := context.Background()

	err = clientmodel.CreateAndAssociate(
		ctx, pool,
		clientName, req.ClientEmail,
		hash,
		resp.CartID,
	)
	if err != nil {
		util.LogError(c, "failed to create client from checkout",
			"error", err,
			"email", req.ClientEmail,
			"cart_id", resp.CartID,
		)
		return
	}

	util.LogInfo(c, "client created from checkout",
		"email", req.ClientEmail,
		"cart_id", resp.CartID,
	)

	if m == nil {
		return
	}

	loginURL, _ := env.SMTPLoginLink()
	resetURL, _ := env.SMTPResetLink()

	if loginURL == "" || resetURL == "" {
		util.LogWarn(c, "mail: login link or reset link not configured, skipping welcome email",
			"email", req.ClientEmail,
		)
		return
	}

	if err := m.SendWelcome(
		ctx, req.ClientEmail,
		clientName, plaintext,
		loginURL, resetURL,
	); err != nil {
		util.LogError(c, "failed to send welcome email",
			"error", err,
			"email", req.ClientEmail,
		)
	} else {
		util.LogInfo(c, "welcome email sent",
			"email", req.ClientEmail,
		)
	}
}
