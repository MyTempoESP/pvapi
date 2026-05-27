package bankaccountmodel

import (
	"context"
	"time"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type BankAccountRow struct {
	ID              int       `db:"id"`
	ShopID          int       `db:"shop_id"`
	Name            string    `db:"name"`
	BankName        string    `db:"bank_name"`
	PixKey          *string   `db:"pix_key"`
	PixTitular      *string   `db:"pix_titular"`
	TransferDetails []byte    `db:"transfer_details"`
	CreatedAt       time.Time `db:"created_at"`
}

func Create(
	ctx context.Context,
	base db.Queryable,
	shopID int,
	req CreateBankAccountRequest,
) (int, error) {
	var transferJSON any
	if req.TransferDetails != nil {
		transferJSON = []byte(*req.TransferDetails)
	}

	return model.Create(
		ctx,
		base,
		query.CreateBankAccountQuery,
		shopID,
		req.Name,
		req.BankName,
		req.PixKey,
		req.PixTitular,
		transferJSON,
	)
}

func ListByShop(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]BankAccountRow, error) {
	return util.WithCollect[BankAccountRow](
		ctx, base,
		query.ListBankAccountsByShopQuery,
		shopID,
	)
}

func FindByID(
	ctx context.Context,
	base db.Queryable,
	id int,
) (BankAccountRow, error) {
	return util.WithCollectExactlyOneRow[BankAccountRow](
		ctx, base,
		query.FindBankAccountByIDQuery,
		id,
	)
}

func Update(
	ctx context.Context,
	base db.Queryable,
	id int,
	req UpdateBankAccountRequest,
) error {
	_, err := base.Exec(
		ctx,
		query.UpdateBankAccountQuery,
		id,
		req.Name,
		req.BankName,
		req.PixKey,
		req.PixTitular,
		req.TransferDetails,
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
		query.DeleteBankAccountQuery,
		id,
	)
	return err
}
