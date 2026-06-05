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
			"qr_code":   resp.QrCode,
		})

		// Best-effort email after response is sent
		if req.ClientEmail != "" {
			sendConfirmationAfterCheckout(c, pool, m, req, resp)
		}
	}
}

func sendConfirmationAfterCheckout(
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

	isNew, err := clientmodel.CreateAndAssociateByCheckout(
		ctx, pool,
		clientName, req.ClientEmail,
		hash,
		resp.CheckoutID,
	)
	if err != nil {
		util.LogError(c, "failed to create client from checkout",
			"error", err,
			"email", req.ClientEmail,
			"checkout_id", resp.CheckoutID,
		)
		return
	}

	util.LogInfo(c, "client created/reused from checkout",
		"email", req.ClientEmail,
		"checkout_id", resp.CheckoutID,
		"is_new", isNew,
	)

	if m == nil {
		return
	}

	// Resolve pickup groups for the email
	groups, err := checkoutmodel.ResolvePickupEmailData(
		ctx, pool,
		req.Items,
	)
	if err != nil {
		util.LogError(c, "failed to resolve pickup email data",
			"error", err,
		)
		return
	}

	emailData := mail.PickupEmailData{
		OrderID:   resp.UniqueID,
		QrCode:    resp.QrCode,
		Groups:    groups,
		IsNewUser: isNew,
		Name:      clientName,
		Email:     req.ClientEmail,
	}

	if isNew {
		loginURL, loginErr := env.SMTPLoginLink()
		resetURL, resetErr := env.SMTPResetLink()
		if loginErr != nil || resetErr != nil || loginURL == "" || resetURL == "" {
			util.LogWarn(c, "mail: login link or reset link not configured, skipping login details",
				"email", req.ClientEmail,
			)
		} else {
			emailData.Password = plaintext
			emailData.LoginURL = loginURL
			emailData.ResetURL = resetURL
		}
	}

	if err := m.SendPickupConfirmation(
		ctx, req.ClientEmail,
		emailData,
	); err != nil {
		util.LogError(c, "failed to send pickup confirmation email",
			"error", err,
			"email", req.ClientEmail,
		)
	} else {
		util.LogInfo(c, "pickup confirmation email sent",
			"email", req.ClientEmail,
		)
	}
}
