package main

import (
	"log/slog"
	"os"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(); err != nil {
		logger.Error(
			"The API encountered a fatal error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() (err error) {
	t := gin.Default()

	// TODO: Check environment variables are properly set

	pool, err := dbNewPool()

	if err != nil {
		return
	}
	defer pool.Close()

	params, err := generateJwtInitParams(pool)

	if err != nil {
		return
	}

	auth, err := jwt.New(params)

	if err != nil {
		return
	}

	err = auth.MiddlewareInit()

	if err != nil {
		return
	}

	registerAuthRoutes(t, auth)

	err = t.Run()

	return
}
