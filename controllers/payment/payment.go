// +build !dev

package payment

import (
	"github.com/gin-gonic/gin"
)

func ValidateTemporarySession(c *gin.Context) {
	c.JSON(200, gin.H{"success": true, "message": "sessao valida"})
}

