package shopmodel

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type ShopCheckoutRow struct {
	ID          int             `db:"checkout_id"`
	UniqueID    string          `db:"checkout_unique_id"`
	Description string          `db:"checkout_description"`
	Value       decimal.Decimal `db:"checkout_value"`
	Tax         decimal.Decimal `db:"checkout_tax"`
	Commission  decimal.Decimal `db:"checkout_commission"`
}

type SaqueRow struct {
	ID              int             `db:"id"`
	Description     string          `db:"description"`
	AskedValue      decimal.Decimal `db:"asked_value"`
	BankName        string          `db:"bank_name"`
	PixCPF          *string         `db:"pix_cpf"`
	PixTitular      *string         `db:"pix_titular"`
	TransferDetails []byte          `db:"transfer_details"`
	Status          string          `db:"status"`
	CreatedAt       time.Time       `db:"created_at"`
	UserName        string          `db:"user_name"`
	ShopName        string          `db:"shop_name"`
}

func BuildCheckoutReport(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) (ShopCheckoutReport, error) {
	var report ShopCheckoutReport

	rows, err := util.WithCollect[ShopCheckoutRow](
		ctx, base,
		query.ShopCheckoutsQuery,
		shopID,
	)
	if err != nil {
		return report, err
	}

	saqueRows, err := util.WithCollect[SaqueRow](
		ctx, base,
		query.ListSaquesByShopQuery,
		shopID,
	)
	if err != nil {
		return report, err
	}

	checkouts := util.ResourceCollection(rows)
	saques := util.ResourceCollection(saqueRows)

	taxes := decimal.NewFromInt(0)
	commissions := decimal.NewFromInt(0)
	total := decimal.NewFromInt(0)

	for _, c := range rows {
		taxes = taxes.Add(c.Tax)
		commissions = commissions.Add(c.Commission)
		total = total.Add(c.Value)
	}

	saquesTotal := decimal.NewFromInt(0)
	for _, s := range saqueRows {
		if s.Status == "approved" {
			saquesTotal = saquesTotal.Add(s.AskedValue)
		}
	}

	actual := total.Sub(taxes).Sub(commissions).Sub(saquesTotal)

	report.Checkouts = checkouts
	report.Saques = saques
	report.Totals = TotalsResource{
		Taxes:       taxes,
		Commissions: commissions,
		Total:       total,
		Saques:      saquesTotal,
		Actual:      actual,
	}

	return report, nil
}
