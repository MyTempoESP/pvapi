package util

import (
	"github.com/gin-gonic/gin"
)

func AbortWithInternalError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(
		500,
		gin.H{
			"status":  false,
			"message": err.Error(), //FIXME:debug (error leak)
		},
	)
}
