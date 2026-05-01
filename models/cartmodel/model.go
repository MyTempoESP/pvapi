package cartmodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/productmodel"
	"github.com/shopspring/decimal"
)

type Cart struct {
	ID        int    `db:"id"`
	ClientID  *int   `db:"client_id"`
	SessionID string `db:"session_id"`
}

type CartItem struct {
	ID 	  int `db:"id"`
	ProductID int `db:"product_id"`
	CartID    int `db:"cart_id"`
	Quantity  int `db:"quantity"`

	Personalizations string `db:"personalizations"` 
}

type CartProduct struct {
	ProductID int    `db:"cart_product_id"`
	Name      string `db:"cart_product_name"`
	Image     string `db:"cart_product_image"`

	ShopID   int    `db:"cart_product_shop"`
	ShopName string `db:"cart_product_shop_name"`

	Price    decimal.Decimal `db:"cart_product_price"`
	Discount decimal.Decimal `db:"cart_product_discount"`

	CartItem
}

type CartDetails struct {
	Items []CartProduct
	Cart
}

func Create(
	ctx context.Context,
	tx pgx.Tx,
	sessionID string,
) error {
	_, err := tx.Exec(ctx, query.CreateCartQuery, sessionID)
	return err
}

func FindBySession(
	ctx context.Context,
	base db.Queryable,
	sessionID string,
) (details CartDetails, err error) {
	cart, err := util.WithCollectExactlyOneRow[Cart](
		ctx,
		base,
		query.CartFindBySession,
		sessionID,
	)

	if err != nil {
		return
	}

	items, err := util.WithCollect[CartProduct](
		ctx,
		base,
		query.ListCartProducts,
		cart.ID,
	)

	if err != nil {
		return
	}

	details.Items = items
	details.Cart = cart

	return
}

func (r AddToCartRequest) Exec(
	ctx context.Context,
	base pgx.Tx,
) (err error) {
	_, err = base.Exec(
		ctx,
		query.AddProductToCartQuery,
		util.Args{
			"quantity": r.Quantity,
			"product_id": r.ProductID,
			"session_id": r.SessionID,
			"personalizations": r.Personalization,
		},
	)
	return
}

func RemoveCartProduct(
	ctx context.Context,
	tx pgx.Tx,
	cartProduct int,
) error {
	_, err := tx.Exec(
		ctx,
		query.RemoveCartProductQuery,
		cartProduct,
	)
	return err
}

func ItemNewQuantity(
	ctx context.Context,
	tx pgx.Tx,
	item int,
	qty int,
) (err error) {
	var productID int

	err = tx.QueryRow(
		ctx,
		query.GetCartItemProductID,
		item,
	).Scan(&productID)

	if err != nil {
		return
	}

	err = productmodel.TakeFromCart(
		ctx, tx,
		productID,
		item,
		qty,
	)

	if err != nil {
		return
	}

	_, err = tx.Exec(
		ctx,
`
update cart_product
set quantity = $2
where id = $1
`,
		item,
		qty,
	)

	return
}
