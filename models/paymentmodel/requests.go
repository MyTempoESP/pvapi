package paymentmodel

import "linhazero.com/pvapi/address"

/*
{
	"payer_name":"ROdrigo",
	"payer_surname":"mtoenafsa asf",
	"payer_email":"aaaaaaaaa@aaaaaa.ca",
	"session_id":"temp_1776880271_73f5d1fd-f16f-4aeb-b7d1-4a703d191154"
}
*/

// Customer represents the customer info inside shipping_info
type Customer struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	Phone string `json:"phone" binding:"required"`
	CPF   string `json:"cpf" binding:"required"`
}

// ShippingInfo holds shipping details
type ShippingInfo struct {
	Type     string                `json:"type" binding:"required"`
	Cost     float64               `json:"cost" binding:"required"`
	Customer Customer              `json:"customer" binding:"required"`
	Address  address.AddressRequest `json:"address" binding:"required"`
}

// GenericPaymentRequest is the incoming payment payload
type GenericPaymentRequest struct {
	PayerName    string       `json:"payer_name" binding:"required"`
	PayerSurname string       `json:"payer_surname" binding:"required"`
	PayerEmail   string       `json:"payer_email" binding:"required"`
	PayerPhone   string       `json:"payer_phone"`
	SessionID    string       `json:"session_id" binding:"required"`
	ShippingInfo ShippingInfo `json:"shipping_info" binding:"required"`
}
