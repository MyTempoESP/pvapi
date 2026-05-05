package model

import (
	"context"

	"linhazero.com/pvapi/db"
)

func Create(
	ctx context.Context,
	base db.Queryable,
	query string,
	params ...any,
) (int, error) {
	var id int

	err := base.QueryRow(
		ctx,
		query,
		params...,
	).Scan(&id)

	return id, err
}

func HasResource(
	ctx context.Context,
	base db.Queryable,
	query string,
	params ...any,
) (hasResource bool, err error) {
	err = base.QueryRow(
		ctx,
		query,
		params...,
	).Scan(&hasResource)

	return
}
