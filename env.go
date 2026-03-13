package main

import (
	"fmt"
	"os"
)

func wrapEnv(k string) func() (string, error) {
	return func() (string, error) {
		v, ok := os.LookupEnv(k)
		if ok {
			return v, nil
		}
		return "", fmt.Errorf("Required environment variable %q not defined", k)
	}
}

var (
	// Auth
	JwtSecretKey = wrapEnv("PV_API_SECRET")

	// Database
	DBHost = wrapEnv("PV_API_DB_HOST")
	DBUser = wrapEnv("PV_API_DB_USER")
	DBPass = wrapEnv("PV_API_DB_PASS")
	DBPort = wrapEnv("PV_API_DB_PORT")
	DBBase = wrapEnv("PV_API_DB_BASE")
)
