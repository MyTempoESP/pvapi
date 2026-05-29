package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/env"
	"linhazero.com/pvapi/mail"
	"linhazero.com/pvapi/middleware/auth"
	"linhazero.com/pvapi/middleware/logmw"
	"linhazero.com/pvapi/seed"
)

func main() {
	handlerOptions := &slog.HandlerOptions{
		Level: &env.LogLevel,
	}

	var logger *slog.Logger

	// XXX
	_, exists := os.LookupEnv("PV_API_PRODUCTION")

	if exists {
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, handlerOptions),
		)
	} else {
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, handlerOptions),
		)
	}

	if err := run(logger); err != nil {
		logger.Error("The API encountered a fatal error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) (err error) {
	// ensuring some essential configs are set
	if _, err = env.DomainName(); err != nil {
		return
	}

	logger.Info("PVAPI: Version 0.54-alpha")

	t := gin.New()
	t.Use(logmw.LogMiddleware(logger))
	t.Use(gin.Recovery())

	t.MaxMultipartMemory = 8 << 20

	// TODO: Check environment variables are properly set

	pool, err := db.NewPool()

	if err != nil {
		return
	}
	defer pool.Close()

	err = seed.Plant(
		logger,
		context.Background(),
		pool,
	)

	if err != nil {
		return
	}

	mailer := createMailer(logger)

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

	setupRoutes(t, authMiddleware, pool, mailer)

	err = t.Run()

	return
}

func createMailer(logger *slog.Logger) mail.Mailer {
	host, err := env.SMTPHost()
	if err != nil {
		// PV_API_SMTP_HOST not set — mail silently disabled
		return nil
	}

	portStr, err := env.SMTPPort()
	if err != nil {
		logger.Error("PV_API_SMTP_HOST is set but PV_API_SMTP_PORT is missing")
		return nil
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Error("PV_API_SMTP_PORT is not a valid number", "value", portStr)
		return nil
	}

	user, err := env.SMTPUser()
	if err != nil {
		logger.Error("PV_API_SMTP_HOST is set but PV_API_SMTP_USER is missing")
		return nil
	}

	pass, err := env.SMTPPass()
	if err != nil {
		logger.Error("PV_API_SMTP_HOST is set but PV_API_SMTP_PASS is missing")
		return nil
	}

	from, err := env.SMTPFrom()
	if err != nil {
		logger.Error("PV_API_SMTP_HOST is set but PV_API_SMTP_FROM is missing")
		return nil
	}

	cfg := mail.Config{
		Host:       host,
		Port:       port,
		User:       user,
		Pass:       pass,
		From:       from,
		SkipVerify: env.SMTPSkipVerify,
	}

	client, err := mail.NewClient(cfg)
	if err != nil {
		logger.Error("failed to create SMTP client", "error", err)
		return nil
	}

	logger.Info("SMTP mail client initialized",
		"host", host,
		"port", port,
	)

	return client
}
