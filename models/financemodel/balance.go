package financemodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/saquemodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type saqueTotalRow struct {
	Total  decimal.Decimal `db:"total"`
	Status string          `db:"status"`
}

type checkoutEarningsRow struct {
	Total decimal.Decimal `db:"total"`
}

type SaqueTotals struct {
	Approved decimal.Decimal
	Pending  decimal.Decimal
}

// AvailableBalance computes the available balance for a given shop.
// Locks the shop row (SELECT … FOR UPDATE) to serialize concurrent
// balance-mutating operations. Must be called inside a transaction.
func AvailableBalance(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) (decimal.Decimal, error) {
	// Lock the shop row
	_, err := base.Exec(ctx, query.LockShopRowQuery, shopID)
	if err != nil {
		return decimal.Zero, err
	}

	// Sum earnings from confirmed checkouts
	earnings, err := util.WithCollectExactlyOneRow[checkoutEarningsRow](
		ctx, base,
		query.CheckoutEarningsByShopQuery,
		shopID,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, err
	}

	// Sum non-terminal saques (pending + approved)
	saqueRows, err := util.WithCollect[saqueTotalRow](
		ctx, base,
		query.SaqueTotalsByShopQuery,
		shopID,
	)
	if err != nil {
		return decimal.Zero, err
	}

	balance := earnings.Total

	for _, s := range saqueRows {
		balance = balance.Sub(s.Total)
	}

	return balance, nil
}

// ComputeSaqueTotals aggregates saque rows into approved and pending sums.
// Pure function — no database calls.
func ComputeSaqueTotals(rows []saquemodel.SaqueRow) SaqueTotals {
	t := SaqueTotals{
		Approved: decimal.NewFromInt(0),
		Pending:  decimal.NewFromInt(0),
	}

	for _, s := range rows {
		switch s.Status {
		case "approved":
			t.Approved = t.Approved.Add(s.AskedValue)
		case "pending":
			t.Pending = t.Pending.Add(s.AskedValue)
		}
	}

	return t
}

// Sum returns the combined sum of approved and pending saques.
func (st SaqueTotals) Sum() decimal.Decimal {
	return st.Approved.Add(st.Pending)
}
