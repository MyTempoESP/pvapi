package clientmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

// SaleRow is a single flat row from ClientSalesQuery.
type SaleRow struct {
	CheckoutID    int             `db:"checkout_id"`
	ShopID        int             `db:"shop_id"`
	ClientID      int             `db:"client_id"`
	ShopValue     decimal.Decimal `db:"shop_value"`
	CreatedAt     time.Time       `db:"created_at"`
	Quantity      int             `db:"quantity"`
	PackageStatus string          `db:"package_status"`
	ProductName   string          `db:"product_name"`
}

// SaleResource is the JSON response shape for one sale.
type SaleResource struct {
	ID       string          `json:"id"`
	ClientID int             `json:"clientId"`
	Products []string        `json:"products"`
	Total    decimal.Decimal `json:"total"`
	Date     string          `json:"date"`
	Status   string          `json:"status"`
}

// saleGroup accumulates data for one checkout_shop composite key.
type saleGroup struct {
	ClientID   int
	Total      decimal.Decimal
	Date       time.Time
	products   map[string]int // product name → quantity
	anyPending bool           // true if any package is not 'recebido'
}

// saleKey is the composite key (checkout_id, shop_id).
type saleKey struct {
	CheckoutID int
	ShopID     int
}

// ListClientSales fetches sales for a client scoped to a shop.
func ListClientSales(
	ctx context.Context,
	pool *pgxpool.Pool,
	clientID int,
	shopID int,
) ([]SaleResource, error) {
	rows, err := util.WithCollect[SaleRow](
		ctx, pool,
		query.ClientSalesQuery,
		clientID,
		shopID,
	)
	if err != nil {
		return nil, fmt.Errorf("query client sales: %w", err)
	}

	groups := make(map[saleKey]*saleGroup)
	keys := make([]saleKey, 0)

	for _, row := range rows {
		key := saleKey{CheckoutID: row.CheckoutID, ShopID: row.ShopID}
		g, ok := groups[key]
		if !ok {
			g = &saleGroup{
				ClientID: row.ClientID,
				Total:    row.ShopValue,
				Date:     row.CreatedAt,
				products: make(map[string]int),
			}
			groups[key] = g
			keys = append(keys, key)
		}

		if row.ProductName != "" {
			g.products[row.ProductName] += row.Quantity
		}

		if row.PackageStatus != "recebido" {
			g.anyPending = true
		}
	}

	// The first row for each key carries the date and total — but they're
	// the same for all rows with the same key, so it's fine.

	resources := make([]SaleResource, 0, len(keys))
	for _, key := range keys {
		g := groups[key]

		productNames := make([]string, 0, len(g.products))
		for name, qty := range g.products {
			productNames = append(productNames, fmt.Sprintf("%dx %s", qty, name))
		}
		sortStrings(productNames)

		resources = append(resources, SaleResource{
			ID:       fmt.Sprintf("%d-%d", key.CheckoutID, key.ShopID),
			ClientID: g.ClientID,
			Total:    g.Total,
			Date:     g.Date.Format("2006-01-02T15:04:05.000Z07:00"),
			Status:   deriveStatus(g),
			Products: productNames,
		})
	}

	return resources, nil
}

// deriveStatus returns "Concluída" when all packages are 'recebido',
// otherwise "Pendente".
func deriveStatus(g *saleGroup) string {
	if g.anyPending {
		return "Pendente"
	}
	return "Concluída"
}

// sortStrings sorts a slice of strings in place (simple insertion sort for
// small slices, avoiding a sort import).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// ClientIDByUserID returns the client ID for a given user ID, or an error if
// the user does not have a client record.
func ClientIDByUserID(ctx context.Context, pool *pgxpool.Pool, userID int) (int, error) {
	var clientID int
	err := pool.QueryRow(ctx, query.FindClientByUserQuery, userID).Scan(&clientID)
	return clientID, err
}
