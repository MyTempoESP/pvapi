// log modern warfare
package logmw

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const loggerKey = "pv_api_logger"

func LogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		logger := logger.With(
			slog.String("path", c.FullPath()),
			slog.String("query", query),
			slog.String("method", c.Request.Method),
		)

		c.Set(loggerKey, logger)

		c.Next()

		logger.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("body_size", c.Writer.Size()),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("request error", slog.String("error", err.Error()))
			}
		}
	}
}

func GetLogger(c *gin.Context) *slog.Logger {
	if val, exists := c.Get(loggerKey); exists {
		if logger, ok := val.(*slog.Logger); ok {
			return logger
		}
	}
	return slog.Default()
}
