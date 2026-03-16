package env

import (
	"fmt"
	"log/slog"
	"os"
)

func wrapEnv(k string) func() (string, error) {
	return func() (string, error) {
		v, ok := os.LookupEnv(k)
		if ok {
			return v, nil
		}
		return "", fmt.Errorf(
			"Required environment variable %q not defined", k)
	}
}

func toLogLevel(
	logLevelEnv func() (string, error),
) (v slog.LevelVar) {
	v.Set(slog.LevelInfo)

	level, err := logLevelEnv()

	if err != nil {
		return
	}

	if level == "debug" {
		v.Set(slog.LevelDebug)
	}

	return
}

var (
	// Auth
	JwtSecretKey  = wrapEnv("PV_API_SECRET")
	SecureCookies = wrapEnv("PV_API_SECURE_COOKIE")

	// Database
	DBHost = wrapEnv("PV_API_DB_HOST")
	DBUser = wrapEnv("PV_API_DB_USER")
	DBPass = wrapEnv("PV_API_DB_PASS")
	DBPort = wrapEnv("PV_API_DB_PORT")
	DBBase = wrapEnv("PV_API_DB_BASE")

	// Admin user
	AdminUser = wrapEnv("PV_API_ADMIN_USER")
	AdminPass = wrapEnv("PV_API_ADMIN_PASS")
	AdminMail = wrapEnv("PV_API_ADMIN_MAIL")

	// Log level
	LogLevel = toLogLevel(wrapEnv("PV_API_LOG_LEVEL"))
)
