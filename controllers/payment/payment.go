package payment

import (
	"fmt"
	"time"
	"net/http"
	"errors"
	"context"
	"strings"
	"strconv"

	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/models/cartmodel"
	"linhazero.com/pvapi/models/paymentmodel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HandleStatusCheck(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func (c *gin.Context) {
		paymentID := c.Params.ByName("payment")

		if paymentID == "" {
			util.HandleBadRequest(c, errors.New("no paymentID provided"))
			return
		}

		paymentDetails, err := paymentmodel.CheckStatus(
			c.Request.Context(), paymentID,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		if paymentDetails.Status == "processed" {
			err = util.TransactionAware(
				c.Request.Context(), pool,
				func (ctx context.Context, tx pgx.Tx) error {
					return paymentmodel.Confirm(ctx, tx, paymentID)
				},
			)

			if err != nil {
				util.Handle500(c, err)
				return
			}
		}

		c.JSON(200, gin.H{
			"payment_id": paymentDetails.PaymentID,
			"external_referece": paymentDetails.ExternalRef,
			"status": paymentDetails.Status,
			"amount": paymentDetails.AmountF,
			"_amount": paymentDetails.Amount,
		})
	}
}

func CartPIXHandler(
	pool *pgxpool.Pool,
) func (c *gin.Context) {
	return func (c *gin.Context) {
		var request paymentmodel.GenericPaymentRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		// verify if the session is still valid
		// workaround is to create a script to manually
		// unassociate assigned expired session_ids.

		cart, err := cartmodel.FindBySession(
			c.Request.Context(), pool,
			request.SessionID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		if len(cart.Items) == 0 {
			util.HandleBadRequest(c, errors.New("carrinho vazio"))
			return
		}

		payment := request.Prepare(&cart)

		details, err := util.TransactionAwareReturn(
			c.Request.Context(), pool,
			payment.Save,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		// I don't give a single fuck if it's exact or not
		// please go kill yourself
		amountF, _ := details.Amount.Float64()

		c.JSON(200, gin.H{
			"success": true,
			"payment_id": details.PaymentID,
			"message": "pagamento iniciado",
			"amount": amountF, // XXX
			"status": details.Status,
			"qr_code": details.QRCode,
			"qr_code_base64": details.QRCodeBase64,
			"items": util.ResourceCollection(cart.Items),
			"total_items": details.Quantity,
			"cart_type": "session_id",
			"cart_identifier": request.SessionID,
		})
	}
}

func ValidateTemporarySession(
	pool *pgxpool.Pool,
) func (c *gin.Context) {
	return func (c *gin.Context) {
		var request struct{
			SessionID string `json:"session_id"`
		}
		err := c.ShouldBind(&request)
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		sess := request.SessionID

		if !strings.HasPrefix(sess, "temp_") {
			c.JSON(400, gin.H{
				"success": false,
				"error": "sessao invalida, nao possui prefixo temp_",
				"at": sess,
			})
			return
		}

		parts := strings.Split(sess, "_")

		if len(parts) < 3 {
			c.JSON(400, gin.H{
				"success": false,
				"error": "sessao invalida, formato nao reconhecido",
				"at": sess,
			})
			return
		}

		unix, err := strconv.ParseInt(parts[1], 10, 64)

		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error": "sessao invalida, timestamp invalida",
				"at": parts[1],
			})
			return
		}

		since := time.Since(time.Unix(int64(unix), 0))

		if since > (time.Hour * 24 * 7) {
			c.JSON(400, gin.H{
				"success": false,
				"error": "sessao expirada",
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": "sessao valida"})
		return
	}
}

func CreateTemporarySession(
	pool *pgxpool.Pool,
) func (c *gin.Context) {
	return func (c *gin.Context) {
		request := c
		sessionUUID, err := uuid.NewRandom()

		if err != nil {
			util.Handle500(c, err)
			return
		}

		sessionID := fmt.Sprintf(
			"temp_%s_%s",
			strconv.FormatInt(time.Now().Unix(), 10),
			sessionUUID.String(),
		)
		metadata := gin.H{
			"session_type": "temporary",
			"created_at": time.Now().Format(time.RFC3339),
			"expires_at": time.Now().Add(time.Hour * 24),
			"request_ip": request.RemoteIP(),
			"user_agent": request.GetHeader("User-Agent"),
		}

		err = util.TransactionAware(
			c.Request.Context(), pool,
			func(
				ctx context.Context,
				tx pgx.Tx,
			) error {
				return cartmodel.Create(
					ctx, tx, sessionID,
				)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}
	
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"session_id": sessionID,
			"metadata": metadata,
			"message": "Sessao temporaria criada com sucesso",
		})
	}
}

