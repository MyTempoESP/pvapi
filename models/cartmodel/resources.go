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

	PriceWithDiscount    decimal.Decimal `json:"price_with_discount"`
	Subtotal             decimal.Decimal `json:"subtotal"`
	SubtotalWithDiscount decimal.Decimal `json:"subtotal_with_discount"`
}

type CartDetailsResource struct {
	Total decimal.Decimal       `json:"total"`
	Items []CartProductResource `json:"items"`
	Count int                   `json:"item_count"`

	TotalWithDiscount decimal.Decimal `json:"total_with_discount"`
	TotalDiscount     decimal.Decimal `json:"total_discount"`

	SessionID string `json:"session_id"`
}

func (c CartProduct) ToResource(
) CartProductResource {
	quantityDecimal := decimal.NewFromInt(int64(c.Quantity)) // must be sized int
	priceDecrement := c.Discount.Div(decimal.NewFromInt(100)).Mul(c.Price)

	subtotal := c.Price.Mul(quantityDecimal)
	priceWithDiscount := c.Price.Sub(priceDecrement)
	subtotalWithDiscount := priceWithDiscount.Mul(quantityDecimal)

	return CartProductResource{
		ID: c.CartItem.ID,
		ProductID: c.ProductID,
		ShopID: c.ShopID,
		ProductName: c.Name,
		ProductImg: c.Image,
		ShopName: c.ShopName,
		Quantity: c.CartItem.Quantity,
		Price: c.Price,
		Discount: c.Discount,
		PriceWithDiscount: priceWithDiscount,
		Subtotal: subtotal,
		SubtotalWithDiscount: subtotalWithDiscount,
	}
}

func (c CartDetails) ToResource(
	sessionID string,
) CartDetailsResource {
	var (
		total             decimal.Decimal = decimal.NewFromInt(0)
		totalDiscount     decimal.Decimal = decimal.NewFromInt(0)
		totalWithDiscount decimal.Decimal = decimal.NewFromInt(0)
	)

	itemResources := util.ResourceCollection[CartProductResource](c.Items)

	for _, i := range itemResources {
		total.Add(i.Subtotal)
		totalWithDiscount.Add(i.SubtotalWithDiscount)
	}

	totalDiscount = total.Sub(totalWithDiscount)

	return CartDetailsResource{
		Total: total,
		Items: itemResources,
		Count: len(c.Items),
		TotalWithDiscount: totalWithDiscount,
		TotalDiscount: totalDiscount,
		SessionID: sessionID,
	}
}
