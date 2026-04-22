package payment

import (
	"fmt"
	"time"
	"net/http"
	"context"
	"strings"
	"strconv"

	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/models/cartmodel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
