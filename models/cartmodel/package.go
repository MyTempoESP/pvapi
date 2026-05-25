package cartmodel

import (
	"context"

	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type CartPackageProduct struct {
	ProductID   int    `db:"product_id"`
	ProductName string `db:"product_name"`
}

type PackageOption struct {
	ProductID     int             `db:"product_id"`
	PackageID     int             `db:"package_id"`
	PackagePrice  decimal.Decimal `db:"package_price"`
	PackageStock  int             `db:"package_stock"`
	Discount      decimal.Decimal `db:"package_discount"`
	ConsignmentID int             `db:"consignment_id"`
	PickupID      int             `db:"pickup_id"`
	PickupName    string          `db:"pickup_name"`
	PickupDate    string          `db:"pickup_date"`
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

func findCartProducts(
	ctx context.Context,
	base db.Queryable,
	sessionID string,
) ([]CartPackageProduct, error) {
	return util.WithCollect[CartPackageProduct](
		ctx, base,
		query.CartProductsForPackages,
		sessionID,
	)
}

func findAvailablePackages(
	ctx context.Context,
	base db.Queryable,
	sessionID string,
) ([]PackageOption, error) {
	return util.WithCollect[PackageOption](
		ctx, base,
		query.AvailablePackages,
		sessionID,
	)
}

func FindPackageOptions(
	ctx context.Context,
	base db.Queryable,
	sessionID string,
) ([]ProductPackageOptions, error) {
	products, err := findCartProducts(ctx, base, sessionID)
	if err != nil {
		return nil, err
	}

	packages, err := findAvailablePackages(ctx, base, sessionID)
	if err != nil {
		return nil, err
	}

	return groupByProduct(products, packages), nil
}

func groupByProduct(
	products []CartPackageProduct,
	packages []PackageOption,
) []ProductPackageOptions {
	pkgByProduct := make(map[int][]PackageOptionResource)
	for _, pkg := range packages {

		// append new option if already exists
		pkgByProduct[pkg.ProductID] = append(
			pkgByProduct[pkg.ProductID],
			PackageOptionResource{
				PackageID:     pkg.PackageID,
				ConsignmentID: pkg.ConsignmentID,
				Price:         pkg.PackagePrice,
				Discount:      pkg.Discount,
				Stock:         pkg.PackageStock,
				Pickup: PickupResource{
					ID:      pkg.PickupID,
					Name:    pkg.PickupName,
					Date:    pkg.PickupDate,
					Address: pkg.Address,
				},
			},
		)
	}

	result := make([]ProductPackageOptions, 0, len(products))
	for _, p := range products {
		pkgs := pkgByProduct[p.ProductID]

		if len(pkgs) == 0 {
			result = append(result, ProductPackageOptions{
				ProductID:   p.ProductID,
				ProductName: p.ProductName,
				Packages:    []PackageOptionResource{},
			})
			continue
		}

		result = append(result, ProductPackageOptions{
			ProductID:   p.ProductID,
			ProductName: p.ProductName,
			Packages:    pkgs,
		})
	}

	return result
}
