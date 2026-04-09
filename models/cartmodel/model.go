package cartmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type Cart struct {
	ID        int    `db:"id"`
	ClientID  *int   `db:"client_id"` //nullable
	SessionID string `db:"session_id"`
}

func (r AddToCartRequest) Exec(
	ctx context.Context,
	base pgx.Tx,
) (err error) {
	_, err = base.Exec(
		ctx,
		query.AddProductToCartQuery,
		util.Args{
			"quantity":         r.Quantity,
			"product_id":       r.ProductID,
			"session_id":       r.SessionID,
			"personalizations": r.Personalization,
		},
	)
	return
}
