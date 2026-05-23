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

type CartConsignmentOption struct {
	ProductID   int    `db:"product_id"`
	ProductName string `db:"product_name"`

	ConsignmentID       *int             `db:"consignment_id"`
	ConsignmentStatus   *string          `db:"consignment_status"`
	ConsignmentCode     *string          `db:"consignment_code"`
	ConsignmentPrice    *decimal.Decimal `db:"consignment_price"`
	ConsignmentStock    *int             `db:"consignment_stock"`
	ConsignmentDiscount *decimal.Decimal `db:"consignment_discount"`

	PickupID   *int       `db:"pickup_id"`
	PickupName *string    `db:"pickup_name"`
	PickupDate *time.Time `db:"pickup_date"`

	address.Address
}

type ConsignmentOptionResource struct {
	ConsignmentID     int             `json:"consignment_id"`
	PriceInConsignment decimal.Decimal `json:"price_in_consignment"`
	Discount          decimal.Decimal `json:"discount"`
	ConsignmentStock  int             `json:"consignment_stock"`
	Pickup            PickupResource  `json:"pickup"`
}

type PickupResource struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Date    string          `json:"date"`
	Address address.Address `json:"address"`
}

type ProductConsignmentOptions struct {
	ProductID    int                         `json:"product_id"`
	ProductName  string                      `json:"product_name"`
	Consignments []ConsignmentOptionResource `json:"consignments"`
}

func FindConsignmentOptions(
	ctx context.Context,
	base db.Queryable,
	sessionID string,
) ([]CartConsignmentOption, error) {
	return util.WithCollect[CartConsignmentOption](
		ctx, base,
		query.ListCartConsignmentOptions,
		sessionID,
	)
}

func GroupByProduct(rows []CartConsignmentOption) []ProductConsignmentOptions {
	groups := make(map[int]*ProductConsignmentOptions)

	for _, row := range rows {
		group, exists := groups[row.ProductID]
		if !exists {
			group = &ProductConsignmentOptions{
				ProductID:    row.ProductID,
				ProductName:  row.ProductName,
				Consignments: make([]ConsignmentOptionResource, 0),
			}
			groups[row.ProductID] = group
		}

		if row.ConsignmentID == nil {
			continue
		}

		group.Consignments = append(group.Consignments, ConsignmentOptionResource{
			ConsignmentID:     *row.ConsignmentID,
			PriceInConsignment: *row.ConsignmentPrice,
			Discount:          *row.ConsignmentDiscount,
			ConsignmentStock:  *row.ConsignmentStock,
			Pickup: PickupResource{
				ID:      *row.PickupID,
				Name:    *row.PickupName,
				Date:    row.PickupDate.Format("2006-01-02T15:04:05"),
				Address: row.Address,
			},
		})
	}

	result := make([]ProductConsignmentOptions, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	return result
}
