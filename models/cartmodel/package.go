package cartmodel

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type CartPackageOption struct {
	ProductID   int    `db:"product_id"`
	ProductName string `db:"product_name"`

	PackageID       *int             `db:"package_id"`
	PackagePrice    *decimal.Decimal `db:"package_price"`
	PackageStock    *int             `db:"package_stock"`
	PackageDiscount *decimal.Decimal `db:"package_discount"`

	ConsignmentID *int `db:"consignment_id"`

	PickupID   *int       `db:"pickup_id"`
	PickupName *string    `db:"pickup_name"`
	PickupDate *time.Time `db:"pickup_date"`

	address.Address
}

type PackageOptionResource struct {
	PackageID      int             `json:"package_id"`
	ConsignmentID  int             `json:"consignment_id"`
	Price          decimal.Decimal `json:"price"`
	Discount       decimal.Decimal `json:"discount"`
	Stock          int             `json:"stock"`
	Pickup         PickupResource  `json:"pickup"`
}

type PickupResource struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Date    string          `json:"date"`
	Address address.Address `json:"address"`
}

type ProductPackageOptions struct {
	ProductID   int                     `json:"product_id"`
	ProductName string                  `json:"product_name"`
	Packages    []PackageOptionResource `json:"packages"`
}

func FindPackageOptions(
	ctx context.Context,
	base db.Queryable,
	sessionID string,
) ([]CartPackageOption, error) {
	return util.WithCollect[CartPackageOption](
		ctx, base,
		query.ListCartPackageOptions,
		sessionID,
	)
}

func GroupByProduct(rows []CartPackageOption) []ProductPackageOptions {
	groups := make(map[int]*ProductPackageOptions)

	for _, row := range rows {
		group, exists := groups[row.ProductID]
		if !exists {
			group = &ProductPackageOptions{
				ProductID:   row.ProductID,
				ProductName: row.ProductName,
				Packages:    make([]PackageOptionResource, 0),
			}
			groups[row.ProductID] = group
		}

		if row.PackageID == nil {
			continue
		}

		group.Packages = append(group.Packages, PackageOptionResource{
			PackageID:     *row.PackageID,
			ConsignmentID: *row.ConsignmentID,
			Price:         *row.PackagePrice,
			Discount:      *row.PackageDiscount,
			Stock:         *row.PackageStock,
			Pickup: PickupResource{
				ID:      *row.PickupID,
				Name:    *row.PickupName,
				Date:    row.PickupDate.Format("2006-01-02T15:04:05"),
				Address: row.Address,
			},
		})
	}

	result := make([]ProductPackageOptions, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	return result
}
