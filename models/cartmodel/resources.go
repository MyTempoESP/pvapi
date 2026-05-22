package cartmodel

import (
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/util"
)

type CartProductResource struct {
	ID        int `json:"id"`
	ProductID int `json:"product_id"`
	ShopID    int `json:"shop_id"`

	ProductName string `json:"product_name"`
	ProductImg  string `json:"product_image"`
	ShopName    string `json:"shop_name"`

	Quantity int             `json:"quantity"`
	Price    decimal.Decimal `json:"price"`
	Discount decimal.Decimal `json:"discount"`

	// FIXME: bring the json back to normal
	Subtotal             decimal.Decimal `json:"_subtotal"`
	PriceWithDiscount    decimal.Decimal `json:"_price_with_discount"`
	SubtotalWithDiscount decimal.Decimal `json:"_subtotal_with_discount"`

	// XXX: tmp
	SubtotalF             float64 `json:"subtotal"`
	PriceWithDiscountF    float64 `json:"price_with_discount"`
	SubtotalWithDiscountF float64 `json:"subtotal_with_discount"`
}

type CartDetailsResource struct {
	// FIXME: bring the json back to normal
	Total decimal.Decimal `json:"_total"`

	// XXX: tmp
	TotalF float64 `json:"total"`

	Items []CartProductResource `json:"items"`
	Count int                   `json:"item_count"`

	// FIXME: bring the json back to normal
	TotalWithDiscount decimal.Decimal `json:"_total_with_discount"`
	TotalDiscount     decimal.Decimal `json:"_total_discount"`

	// XXX: tmp
	TotalWithDiscountF float64 `json:"total_with_discount"`
	TotalDiscountF     float64 `json:"total_discount"`

	SessionID string `json:"session_id"`
}

func (c CartProduct) ToResource() CartProductResource {
	quantityDecimal := decimal.NewFromInt(int64(c.Quantity)) // must be sized int
	priceDecrement := c.Discount.Div(decimal.NewFromInt(100)).Mul(c.Price)

	subtotal := c.Price.Mul(quantityDecimal)
	priceWithDiscount := c.Price.Sub(priceDecrement)
	subtotalWithDiscount := priceWithDiscount.Mul(quantityDecimal)

	// XXX: remove this
	subtotalF, _ := subtotal.Float64()
	priceWithDiscountF, _ := priceWithDiscount.Float64()
	subtotalWithDiscountF, _ := subtotalWithDiscount.Float64()

	return CartProductResource{
		ID:                   c.ID,
		ProductID:            c.ProductID,
		ShopID:               c.ShopID,
		ProductName:          c.Name,
		ProductImg:           c.Image,
		ShopName:             c.ShopName,
		Quantity:             c.Quantity,
		Price:                c.Price,
		Discount:             c.Discount,
		Subtotal:             subtotal,
		PriceWithDiscount:    priceWithDiscount,
		SubtotalWithDiscount: subtotalWithDiscount,

		// XXX: this is blasphemy atp
		SubtotalF:             subtotalF,
		PriceWithDiscountF:    priceWithDiscountF,
		SubtotalWithDiscountF: subtotalWithDiscountF,
	}
}

func (c CartDetails) ToResource(
	sessionID string,
) CartDetailsResource {
	var (
		total             = decimal.NewFromInt(0)
		totalDiscount     = decimal.NewFromInt(0)
		totalWithDiscount = decimal.NewFromInt(0)
	)

	itemResources := util.ResourceCollection[CartProductResource](c.Items)

	for _, i := range itemResources {
		total.Add(i.Subtotal)
		totalWithDiscount.Add(i.SubtotalWithDiscount)
	}

	totalDiscount = total.Sub(totalWithDiscount)

	// XXX: this makes me so pissed off
	totalWithDiscountF, _ := totalWithDiscount.Float64()
	totalDiscountF, _ := totalDiscount.Float64()

	totalF, _ := total.Float64()

	return CartDetailsResource{
		Total:             total,
		TotalF:            totalF,
		Items:             itemResources,
		Count:             len(c.Items),
		TotalWithDiscount: totalWithDiscount,
		TotalDiscount:     totalDiscount,

		// XXX: remove this
		TotalWithDiscountF: totalWithDiscountF,
		TotalDiscountF:     totalDiscountF,

		SessionID: sessionID,
	}
}
