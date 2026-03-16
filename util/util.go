package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"linhazero.com/pvapi/middleware/logmw"
)

// aliasing for centralization
func HashPassword(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword(
		[]byte(pass), 12,
	)
}

func ComparePasswordWithHash(pass string, hashed []byte) error {
	return bcrypt.CompareHashAndPassword(hashed, []byte(pass))
}

func HandleBadRequest(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Bad request", "error", err)

	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{
			"status":  false,
			"message": err.Error(), //FIXME:debug (error leak)
		},
	)
}

func Handle404(c *gin.Context) {
	logger := logmw.GetLogger(c)
	logger.Warn("Resource not found")

	c.AbortWithStatusJSON(
		404,
		gin.H{
			"message": "resource not found",
		},
	)
}

func Handle500(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Error("Internal server error", "error", err)

	c.AbortWithStatusJSON(
		500,
		gin.H{
			"status":  false,
			"message": err.Error(), //FIXME:debug (error leak)
		},
	)
}
