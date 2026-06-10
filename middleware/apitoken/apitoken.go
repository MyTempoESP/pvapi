package apitoken

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/apitokenmodel"
)

const (
	ContextKeyAPIClientID = "api_client_id"
)

// Require returns a gin middleware that validates an API bearer token.
// Only tokens with the "pv_" prefix are accepted.
// On success, the api_client_id is injected into the gin context.
func Require(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"message": "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"message": "invalid authorization header format",
			})
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")

		if !strings.HasPrefix(token, apitokenmodel.TokenPrefix) {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"message": "invalid token type",
			})
			return
		}

		result, err := apitokenmodel.ValidateToken(c.Request.Context(), pool, token)
		if err != nil {
			switch {
			case errors.Is(err, apitokenmodel.ErrTokenNotFound):
				c.AbortWithStatusJSON(401, gin.H{
					"success": false,
					"message": "invalid token",
				})
			case errors.Is(err, apitokenmodel.ErrTokenRevoked):
				c.AbortWithStatusJSON(401, gin.H{
					"success": false,
					"message": "token revoked",
				})
			default:
				c.AbortWithStatusJSON(500, gin.H{
					"success": false,
					"message": "internal server error",
				})
			}
			return
		}

		c.Set(ContextKeyAPIClientID, result.ID)
		c.Next()
	}
}

// GetAPIClientID extracts the API client ID from the gin context.
// Returns 0 if not set.
func GetAPIClientID(c *gin.Context) int {
	id, exists := c.Get(ContextKeyAPIClientID)
	if !exists {
		return 0
	}
	v, ok := id.(int)
	if !ok {
		return 0
	}
	return v
}
