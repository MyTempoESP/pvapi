package shopmodel

import (
	"context"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/financemodel"
	"linhazero.com/pvapi/models/saquemodel"
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

	saqueRows, err := saquemodel.ListByShop(
		ctx, base,
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

	saqueTotals := financemodel.ComputeSaqueTotals(saqueRows)
	saquesCombined := saqueTotals.Sum()

	actual := total.Sub(taxes).Sub(commissions).Sub(saquesCombined)

	report.Checkouts = checkouts
	report.Saques = saques
	report.Totals = TotalsResource{
		Taxes:       taxes,
		Commissions: commissions,
		Total:       total,
		Saques:      saquesCombined,
		Actual:      actual,
	}

	return report, nil
}
