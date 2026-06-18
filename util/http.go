package util

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/middleware/logmw"
)

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

func Handle404(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Resource not found", "error", err)

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

func LogInfo(c *gin.Context, message string, a ...any) {
	logger := logmw.GetLogger(c)
	logger.Info(message, a...)
}

func LogWarn(c *gin.Context, message string, a ...any) {
	logger := logmw.GetLogger(c)
	logger.Warn(message, a...)
}

func LogError(c *gin.Context, message string, a ...any) {
	logger := logmw.GetLogger(c)
	logger.Error(message, a...)
}

func HandleGuess(c *gin.Context, err error) {
	if errors.Is(err, ErrBadRequest) {
		HandleBadRequest(c, err)
		return
	}

	if errors.Is(err, ErrNotFound) {
		Handle404(c, err)
		return
	}

	Handle500(c, err)
}

func HandleGuessFromDBOperation(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		Handle404(c, err)
		return
	}

	if errors.Is(err, ErrBadRequest) {
		HandleBadRequest(c, err)
		return
	}

	if errors.Is(err, ErrDuplicatedResource) {
		HandleDuplicatedResource(c, err)
		return
	}

	if errors.Is(err, ErrResourceLocked) {
		HandleResourceLocked(c, err)
		return
	}

	if errors.Is(err, ErrNotFound) {
		Handle404(c, err)
		return
	}

	Handle500(c, err)
}

func HandleGone(c *gin.Context, message string) {
	c.AbortWithStatusJSON(
		410,
		gin.H{
			"message": message,
		},
	)
}

func HandleDuplicatedResource(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Duplicated resource", "error", err)

	c.AbortWithStatusJSON(
		http.StatusConflict,
		gin.H{
			"success": false,
			"message": "resource already exists",
			"code":    "duplicated_resource",
		},
	)
}

func HandleResourceLocked(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Resource locked", "error", err)

	c.AbortWithStatusJSON(
		http.StatusConflict,
		gin.H{
			"success": false,
			"message": "não é possível remover: produto possui reservas ativas em checkouts",
			"code":    "resource_locked",
		},
	)
}

func WrapLabeled[R any](
	c *gin.Context,
	message string,
	label string, data R,
) {
	c.JSON(200, gin.H{
		"success": true,
		"message": message,
		label:     data,
	})
}

func Wrap[R any](
	c *gin.Context,
	message string,
	data R,
) {
	WrapLabeled(c, message, "data", data)
}

func WrapMap(
	c *gin.Context,
	message string,
	json gin.H,
) {
	json["success"] = true
	json["message"] = message
	c.JSON(200, json)
}

func OK(
	c *gin.Context,
	message string,
) {
	c.JSON(200, gin.H{
		"success": true,
		"message": message,
	})
}
