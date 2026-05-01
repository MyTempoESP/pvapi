package paymentmodel

/* models/paymentmodel/blueprints.go */

import (
	"linhazero.com/pvapi/models/cartmodel"
)

type PaymentBlueprint struct {
	PayerName    string
	PayerSurname string
	PayerEmail   string
	PayerPhone   string

	Cart *cartmodel.CartDetails
}

func (p GenericPaymentRequest) Prepare(
	cart *cartmodel.CartDetails,
) PaymentBlueprint {
	return PaymentBlueprint{
		PayerName: p.PayerName,
		PayerSurname: p.PayerSurname,
		PayerEmail: p.PayerEmail,
		PayerPhone: p.PayerPhone,
		Cart: cart,
	}
}

