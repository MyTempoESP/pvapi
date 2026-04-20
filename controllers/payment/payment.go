package payment

import (
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateTemporarySession(
	c *gin.Context,
) {
	request := c
	sessionId, _ := uuid.NewRandom()
	metadata := gin.H{
		"session_type": "temporary",
		"created_at": time.Now().Format(time.RFC3339),
		"expires_at": time.Now().Add(time.Hour * 24),
		"request_ip": request.RemoteIP(),
		"user_agent": request.GetHeader("User-Agent"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"session_id": sessionId,
		"metadata": metadata,
		"message": "Sessao temporaria criada com sucesso",
	})
}
