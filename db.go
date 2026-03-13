package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func buildDBUrl() (url string, err error) {
	var host, port, user, pass, base string

	if host, err = DBHost(); err != nil {
		return
	}

	if port, err = DBPort(); err != nil {
		return
	}

	if user, err = DBUser(); err != nil {
		return
	}

	if pass, err = DBPass(); err != nil {
		return
	}

	if base, err = DBBase(); err != nil {
		return
	}

	url = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, base,
	)

	return
}

func dbNewPool() (pool *pgxpool.Pool, err error) {
	var url string

	url, err = buildDBUrl()

	if err != nil {
		return
	}

	pool, err = pgxpool.New(context.Background(), url)

	if err != nil {
		return
	}

	return
}
