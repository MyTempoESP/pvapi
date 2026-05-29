package saquemodel

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type SaqueRow struct {
	ID              int             `db:"id"`
	Description     string          `db:"description"`
	AskedValue      decimal.Decimal `db:"asked_value"`
	BankAccountID   *int            `db:"bank_account_id"`
	BankName        *string         `db:"bank_name"`
	PixKey          *string         `db:"pix_key"`
	PixTitular      *string         `db:"pix_titular"`
	TransferDetails []byte          `db:"transfer_details"`
	Status          string          `db:"status"`
	CreatedAt       time.Time       `db:"created_at"`
	UserName        string          `db:"user_name"`
	ShopName        string          `db:"shop_name"`
}

func Create(
	ctx context.Context,
	base db.Queryable,
	shopID int,
	userID int,
	req CreateSaqueRequest,
) (int, error) {
	return model.Create(
		ctx,
		base,
		query.CreateSaqueQuery,
		shopID,
		userID,
		req.Description,
		req.AskedValue,
		req.BankAccountID,
	)
}

func ListByShop(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]SaqueRow, error) {
	return util.WithCollect[SaqueRow](
		ctx, base,
		query.ListSaquesByShopQuery,
		shopID,
	)
}

func ListAll(
	ctx context.Context,
	base db.Queryable,
) ([]SaqueRow, error) {
	return util.WithCollect[SaqueRow](
		ctx, base,
		query.ListAllSaquesQuery,
	)
}

func FindByID(
	ctx context.Context,
	base db.Queryable,
	id int,
) (SaqueRow, error) {
	return util.WithCollectExactlyOneRow[SaqueRow](
		ctx, base,
		query.FindSaqueByIDQuery,
		id,
	)
}

func UpdateStatus(
	ctx context.Context,
	base db.Queryable,
	id int,
	status string,
) error {
	_, err := base.Exec(
		ctx,
		query.UpdateSaqueStatusQuery,
		id,
		status,
	)
	return err
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id int,
) error {
	_, err := base.Exec(
		ctx,
		query.DeleteSaqueQuery,
		id,
	)
	return err
}
