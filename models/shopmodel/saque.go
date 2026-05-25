package shopmodel

import (
	"context"
	"encoding/json"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type CreateSaqueRequest struct {
	Description     string           `json:"description" binding:"required"`
	AskedValue      decimal.Decimal  `json:"asked_value" binding:"required"`
	BankName        string           `json:"bank_name" binding:"required"`
	PixCPF          *string          `json:"pix_cpf"`
	PixTitular      *string          `json:"pix_titular"`
	TransferDetails *json.RawMessage `json:"transfer_details"`
}

func CreateSaque(
	ctx context.Context,
	base db.Queryable,
	shopID int,
	userID int,
	req CreateSaqueRequest,
) (int, error) {
	var transferJSON any
	if req.TransferDetails != nil {
		transferJSON = []byte(*req.TransferDetails)
	}

	var id int
	err := base.QueryRow(
		ctx,
		query.CreateSaqueQuery,
		shopID,
		userID,
		req.Description,
		req.AskedValue,
		req.BankName,
		req.PixCPF,
		req.PixTitular,
		transferJSON,
	).Scan(&id)

	return id, err
}

func ListSaquesByShop(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]SaqueResource, error) {
	rows, err := util.WithCollect[SaqueRow](
		ctx, base,
		query.ListSaquesByShopQuery,
		shopID,
	)
	if err != nil {
		return nil, err
	}

	return util.ResourceCollection(rows), nil
}
