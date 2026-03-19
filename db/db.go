package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/env"
)

type Queryable interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func buildDBUrl() (url string, err error) {
	var host, port, user, pass, base string

	if host, err = env.DBHost(); err != nil {
		return
	}

	if port, err = env.DBPort(); err != nil {
		return
	}

	if user, err = env.DBUser(); err != nil {
		return
	}

	if pass, err = env.DBPass(); err != nil {
		return
	}

	if base, err = env.DBBase(); err != nil {
		return
	}

	url = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, base,
	)

	return
}

func NewPool() (pool *pgxpool.Pool, err error) {
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
