package util

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/db"
)

var SQL = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// A is a quick alias for NamedArgs
type Args = pgx.NamedArgs

func GetPgError(err error) (pge *pgconn.PgError, ok bool) {
	ok = errors.As(err, &pge)
	return
}

func WithCollectExactlyOneRow[T any](
	ctx context.Context,
	base db.Queryable,
	query string,
	params ...any,
) (T, error) {
	var zero T

	row, err := base.Query(ctx, query, params...)
	if err != nil {
		return zero, err
	}

	return pgx.CollectExactlyOneRow(
		row, pgx.RowToStructByName[T],
	)
}

func WithCollect[T any](
	ctx context.Context,
	base db.Queryable,
	query string,
	params ...any,
) ([]T, error) {
	rows, err := base.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(
		rows, pgx.RowToStructByName[T],
	)
}

func Bake(
	ctx context.Context,
	base db.Queryable,
	batch *pgx.Batch,
) (err error) {
	b := base.SendBatch(ctx, batch)
	defer func() { _ = b.Close() }()

	for range batch.Len() {
		_, err = b.Exec()
		if err != nil {
			return
		}
	}

	err = b.Close()
	return
}

func Transaction(
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, db.Queryable) error,
) error {
	return pgx.BeginFunc(ctx, pool,
		func(tx pgx.Tx) error {
			return f(ctx, tx)
		},
	)
}

func TransactionAware(
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, pgx.Tx) error,
) error {
	return pgx.BeginFunc(ctx, pool,
		func(tx pgx.Tx) error {
			return f(ctx, tx)
		},
	)
}

func TransactionReturn[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, db.Queryable) (T, error),
) (T, error) {
	var zero T

	tx, err := pool.Begin(ctx)
	if err != nil {
		return zero, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	t, err := f(ctx, tx)
	if err != nil {
		return zero, err
	}

	err = tx.Commit(ctx)
	return t, err
}

func TransactionAwareReturn[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, pgx.Tx) (T, error),
) (T, error) {
	var zero T

	tx, err := pool.Begin(ctx)
	if err != nil {
		return zero, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	t, err := f(ctx, tx)
	if err != nil {
		return zero, err
	}

	err = tx.Commit(ctx)
	return t, err
}
