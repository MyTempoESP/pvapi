package main

import (
	"log/slog"
	"os"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"linhazero.com/pvapi/auth"
	"linhazero.com/pvapi/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(); err != nil {
		logger.Error("The API encountered a fatal error", "error", err)
		os.Exit(1)
	}
}

func run() (err error) {
	t := gin.Default()

	// TODO: Check environment variables are properly set

	pool, err := db.NewPool()

	if err != nil {
		return
	}
	defer pool.Close()

	params, err := auth.GenerateJwtInitParams(pool)

	if err != nil {
		return
	}

	authMiddleware, err := jwt.New(params)

	if err != nil {
		return
	}

	err = authMiddleware.MiddlewareInit()

	if err != nil {
		return
	}

	setupRoutes(t, authMiddleware, pool)

	err = t.Run()

	return
}
