package consignmentmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func CreateTx(
	ctx context.Context,
	tx pgx.Tx,
	consignment CreateConsignmentRequest,
) (err error) {

	return
}
