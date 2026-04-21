package payment

import (
	"time"
	"net/http"
	"context"

	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/models/cartmodel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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

		sessionID := sessionUUID.String()
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
