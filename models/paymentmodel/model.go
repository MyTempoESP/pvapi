package paymentmodel

/* models/paymentmodel/model.go */


import (

	"errors"
	"time"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/order"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/env"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/models/cartmodel"
	"linhazero.com/pvapi/models/clientmodel"
)


type Payment struct {
	ID     int
	CartID int

	PaymentID    string
	ExternalRef  string
	Title        string
	Description  string
	Amount       decimal.Decimal
	Quantity     int
	PayerEmail   string
	PayerName    string
	PayerSurname string
	Status       string
	Method       string
	TAmount      decimal.Decimal
	Currency     string
	ProcessedAt  time.Time
}

type PIXMetadata struct {
	QRCode       string
	QRCodeBase64 string
}
type PIXPayment struct {
	Payment
	PIXMetadata `db:"metadata"`
}

func percentageOf(total, pct decimal.Decimal) decimal.Decimal {
	return total.Mul(pct.Div(decimal.NewFromInt(100)))
}

func calculateItemSubtotal(item cartmodel.CartProduct) decimal.Decimal {
	/* get the actual money discount */
	discountedAmount := percentageOf(item.Price, item.Discount)
	discountedPrice := item.Price.Sub(discountedAmount)

	/* multiplying price with the quantity; NewFromInt expects int64 */
	return discountedPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
}

func toPIXRequest(
	payment PaymentBlueprint,
) (order.Request, error) {
	var zero order.Request

	/* calculate the totals */
	total := decimal.NewFromInt(0)
	for _, item := range(payment.Cart.Items) {
		total = total.Add(calculateItemSubtotal(item))
	}

	if total.LessThanOrEqual(decimal.Zero) {
		return zero, fmt.Errorf("invalid cart total: %s", total.String())
	}

	/* TODO: this is too simple */
	description := fmt.Sprintf(
		"Compra de %d produto(s)", len(payment.Cart.Items),
	)

	/* external Reference as cart_id + session_id */
	reference := fmt.Sprintf(
		"CART_PIX_%d_%s",
		payment.Cart.ID,
		payment.Cart.SessionID,
	)

	/* formatting amount doing the so called 'bank-rounding'  */
	amount := total.StringFixedBank(2)

	return order.Request{
		Type: "online",
		TotalAmount: amount,
		ExternalReference: reference,
		Currency: "BRL",
		Description: description,
		Transactions: &order.TransactionRequest{
			Payments: []order.PaymentRequest{
				{
				Amount: amount,
				PaymentMethod: &order.PaymentMethodRequest{
					ID: "pix",
					Type: "bank_transfer",
				},
				},
			},
		},
		Payer: &order.PayerRequest{
			FirstName: payment.PayerName,
			LastName: payment.PayerSurname,
			Email: payment.PayerEmail,
		},
	}, nil
}

func newMercadoPago() (order.Client, error) {
	accessToken, err := env.PaymentToken()

	if err != nil {
		return nil, err
	}

	c, err := config.New(accessToken)

	if err != nil {
		return nil, err
	}

	return order.NewClient(c), nil
}

func Create(
	ctx context.Context,
	tx pgx.Tx,
	payment PIXPayment,
) error {
	_, err := tx.Exec(
		ctx,
		query.PaymentCreateQuery,
		util.Args{
			"cart_id": payment.CartID,
			"transaction_amount": payment.TAmount,
			"quantity": payment.Quantity,
			"amount": payment.Amount,
			"status": payment.Status,
			"description": payment.Description,
			"title": payment.Title,
			"payment_id": payment.PaymentID,
			"payer_name": payment.PayerName,
			"payer_email": payment.PayerEmail,
			"payer_surname": payment.PayerSurname,
			"payment_method": payment.Method,
			"currency": payment.Currency,
			"external_reference": payment.ExternalRef,
			"method": payment.Method,
			"metadata": payment.PIXMetadata,
		},
	)

	return err
}

func (p PaymentBlueprint) Save(
	ctx context.Context,
	tx pgx.Tx,
) (prepared PIXPayment, err error) {
	err = clientmodel.CreateForCart(
		ctx, tx,
		p.PayerName,
		p.PayerEmail,
		p.PayerPhone,
		p.Cart.ID,
	)

	if err != nil {
		return
	}

	err = p.Cart.HoldStock(ctx, tx)

	if err != nil {
		return
	}

	paymentClient, err := newMercadoPago()

	if err != nil {
		return
	}

	paymentRequest, err := toPIXRequest(p)

	if err != nil {
		return
	}

	payment, err := paymentClient.Create(ctx, paymentRequest)

	if err != nil {
		err = fmt.Errorf(
			"Mercado pago error: %w, ", err)
		return
	}

	paymentAmount, err := decimal.NewFromString(payment.TotalAmount)

	if err != nil {
		return
	}

	transaction := payment.Transactions

	/* payment processor objects */
	ppos := transaction.Payments

	if len(ppos) < 1 {
		err = errors.New("no payments returned from mercadoPago")
		return
	}

	ppo := ppos[0]

	prepared = PIXPayment{
		Payment: Payment{
			PaymentID: payment.ID,
			CartID: p.Cart.ID,
			ExternalRef: payment.ExternalReference,
			Title: payment.Description,
			Description: payment.Description,
			Amount: paymentAmount,
			Quantity: len(p.Cart.Items),
			PayerName: p.PayerName,
			PayerEmail: p.PayerEmail,
			PayerSurname: p.PayerSurname,
			Status: payment.Status,
			Method: "pix",
			TAmount: paymentAmount,
			Currency: payment.Currency,
			ProcessedAt: time.Now(),
		},
		PIXMetadata: PIXMetadata{
			QRCode: ppo.PaymentMethod.QrCode,
			QRCodeBase64: ppo.PaymentMethod.QrCodeBase64,
		},
	}

	err = Create(ctx, tx, prepared)

	return
}


type PaymentStatus struct {
	PaymentID   string
	ExternalRef string
	Status      string
	AmountF     float64
	Amount      decimal.Decimal
}

func CheckStatus(
	ctx context.Context,
	paymentID string,
) (status PaymentStatus, err error) {
	paymentClient, err := newMercadoPago()

	if err != nil {
		return
	}

	payment, err := paymentClient.Get(ctx, paymentID)

	if err != nil {
		return
	}

	amount, err := decimal.NewFromString(payment.TotalAmount)

	if err != nil {
		return
	}

	amountf, _ := amount.Float64()

	status.PaymentID = paymentID
	status.ExternalRef = payment.ExternalReference
	status.Amount = amount
	status.AmountF = amountf

	if payment.Status == "processed" {
		status.Status = "approved"
	} else {
		status.Status = payment.Status
	}

	return
}


func Confirm(
	ctx context.Context,
	tx pgx.Tx,
	paymentID string,
) (err error) {
	_, err = tx.Exec(
		ctx,
`
update payments
set status = 'processed'
where payment_id = $1
`,
		paymentID,
	)

	return
}
