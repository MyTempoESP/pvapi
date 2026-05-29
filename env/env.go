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
			"required environment variable %q not defined", k)
	}
}

// booleanEnv looks for an environment variable and wheter or not it's set
func booleanEnv(k string) (set bool) {
	_, set = os.LookupEnv(k)
	return
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

	// Domain
	DomainName   = wrapEnv("PV_API_DOMAIN_NAME")
	HTTPSEnabled = booleanEnv("PV_API_HTTPS_ENABLED")

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

	// Payment
	PaymentToken = wrapEnv("PV_API_MP_TOKEN")

	// Log level
	LogLevel = toLogLevel(wrapEnv("PV_API_LOG_LEVEL"))

	// Test mode
	TestMode = wrapEnv("PV_API_TEST_MODE")

	// SMTP (mail)
	SMTPHost       = wrapEnv("PV_API_SMTP_HOST")
	SMTPPort       = wrapEnv("PV_API_SMTP_PORT")
	SMTPUser       = wrapEnv("PV_API_SMTP_USER")
	SMTPPass       = wrapEnv("PV_API_SMTP_PASS")
	SMTPFrom       = wrapEnv("PV_API_SMTP_FROM")
	SMTPLoginLink  = wrapEnv("PV_API_LOGIN_LINK")
	SMTPResetLink  = wrapEnv("PV_API_PASSWORD_RESET_LINK")
	SMTPSkipVerify = booleanEnv("PV_API_SMTP_SKIP_VERIFY")
)
