package paymentmodel

/* models/paymentmodel/blueprints.go */

import (
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/models/cartmodel"
)

type PaymentBlueprint struct {
	PayerName    string
	PayerSurname string
	PayerEmail   string
	PayerPhone   string

	// shipping info
	ShippingType   string
	ShippingCost   float64
	CustomerName   string
	CustomerEmail  string
	CustomerPhone  string
	CustomerCPF    string
	ShippingAddress address.Address

	Cart *cartmodel.CartDetails
}

func (p GenericPaymentRequest) Prepare(
	cart *cartmodel.CartDetails,
) PaymentBlueprint {
	return PaymentBlueprint{
		PayerName:    p.PayerName,
		PayerSurname: p.PayerSurname,
		PayerEmail:   p.PayerEmail,
		PayerPhone:   p.PayerPhone,

		ShippingType:   p.ShippingInfo.Type,
		ShippingCost:   p.ShippingInfo.Cost,
		CustomerName:   p.ShippingInfo.Customer.Name,
		CustomerEmail:  p.ShippingInfo.Customer.Email,
		CustomerPhone:  p.ShippingInfo.Customer.Phone,
		CustomerCPF:    p.ShippingInfo.Customer.CPF,
		ShippingAddress: p.ShippingInfo.Address.Prepare(),

		Cart: cart,
	}
}

