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
	"linhazero.com/pvapi/util/checkoutclient"
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
			"qr_code":   resp.QrCode,
		})

		// Best-effort client creation and email after response is sent
		if ci.EffectiveEmail != "" {
			sendConfirmationAfterCheckout(c, pool, m, ci, req.Items, resp)
		}
	}
}

func sendConfirmationAfterCheckout(
	c *gin.Context,
	pool *pgxpool.Pool,
	m mail.Mailer,
	ci checkoutclient.CheckoutClientOutput,
	items []checkoutmodel.ConfirmItem,
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

	if m == nil || ci.IsSyntheticEmail {
		return
	}

	// Resolve pickup groups for the email
	groups, err := checkoutmodel.ResolvePickupEmailData(
		ctx, pool,
		items,
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
		Name:      ci.ClientName,
		Email:     ci.EffectiveEmail,
	}

	if isNew {
		loginURL, loginErr := env.SMTPLoginLink()
		resetURL, resetErr := env.SMTPResetLink()
		if loginErr != nil || resetErr != nil || loginURL == "" || resetURL == "" {
			util.LogWarn(c, "mail: login link or reset link not configured, skipping login details",
				"email", ci.EffectiveEmail,
			)
		} else {
			emailData.Password = ci.Password
			emailData.LoginURL = loginURL
			emailData.ResetURL = resetURL
		}
	}

	if err := m.SendPickupConfirmation(
		ctx, ci.EffectiveEmail,
		emailData,
	); err != nil {
		util.LogError(c, "failed to send pickup confirmation email",
			"error", err,
			"email", ci.EffectiveEmail,
		)
	} else {
		util.LogInfo(c, "pickup confirmation email sent",
			"email", ci.EffectiveEmail,
		)
	}
}
