package clientmodel

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type overviewCountsRow struct {
	TotalOrders  int `db:"total_orders"`
	PendingCount int `db:"pending_count"`
}

type overviewPackageRow struct {
	CheckoutUniqueID string          `db:"checkout_unique_id"`
	ReservedAt       time.Time       `db:"reserved_at"`
	Status           string          `db:"status"`
	ShopValue        decimal.Decimal `db:"shop_value"`
	Quantity         int             `db:"quantity"`

	PickupID       int       `db:"pickup_id"`
	PickupName     string    `db:"pickup_name"`
	PickupDatetime time.Time `db:"pickup_datetime"`

	address.Address

	ProductID   int    `db:"product_id"`
	ProductName string `db:"product_name"`
	ProductCode string `db:"product_code"`
}

type OverviewAddressResource struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	Neigh      string `json:"neighborhood"`
	City       string `json:"city"`
	State      string `json:"state"`
	CEP        string `json:"cep"`
	Complement string `json:"complement"`
}

type OverviewPickupResource struct {
	ID       int                     `json:"id"`
	Name     string                  `json:"name"`
	Datetime string                  `json:"datetime"`
	Address  OverviewAddressResource `json:"address"`
}

type OverviewProductResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type OverviewPackageResource struct {
	UUID     string                  `json:"uuid"`
	Date     string                  `json:"date"`
	Status   string                  `json:"status"`
	Value    decimal.Decimal         `json:"value"`
	Quantity int                     `json:"quantity"`
	Pickup   OverviewPickupResource  `json:"pickup"`
	Product  OverviewProductResource `json:"product"`
}

type ClientOverview struct {
	Orders   int                       `json:"orders"`
	Pending  int                       `json:"pending"`
	Packages []OverviewPackageResource `json:"packages"`
}

func (r overviewPackageRow) ToResource() OverviewPackageResource {
	addr := OverviewAddressResource{
		Street:     r.Street,
		Number:     r.Number,
		Neigh:      r.Neigh,
		City:       r.City,
		State:      r.State,
		CEP:        r.CEP,
		Complement: r.Complement,
	}

	pickup := OverviewPickupResource{
		ID:       r.PickupID,
		Name:     r.PickupName,
		Datetime: r.PickupDatetime.Format("2006-01-02T15:04:05"),
		Address:  addr,
	}

	product := OverviewProductResource{
		ID:   r.ProductID,
		Name: r.ProductName,
		Code: r.ProductCode,
	}

	return OverviewPackageResource{
		UUID:     r.CheckoutUniqueID,
		Date:     r.ReservedAt.Format("2006-01-02T15:04:05"),
		Status:   r.Status,
		Value:    r.ShopValue,
		Quantity: r.Quantity,
		Pickup:   pickup,
		Product:  product,
	}
}

func GetOverview(
	ctx context.Context,
	pool *pgxpool.Pool,
	clientID int,
) (ClientOverview, error) {
	var overview ClientOverview

	counts, err := util.WithCollectExactlyOneRow[overviewCountsRow](
		ctx, pool,
		query.ClientOverviewCountsQuery,
		clientID,
	)
	if err != nil {
		return overview, err
	}

	overview.Orders = counts.TotalOrders
	overview.Pending = counts.PendingCount

	packages, err := util.WithCollect[overviewPackageRow](
		ctx, pool,
		query.ClientOverviewPackagesQuery,
		clientID,
	)
	if err != nil {
		return overview, err
	}

	overview.Packages = util.ResourceCollection(packages)

	return overview, nil
}
