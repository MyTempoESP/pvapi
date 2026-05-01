package paymentmodel

/*
{
	"payer_name":"ROdrigo",
	"payer_surname":"mtoenafsa asf",
	"payer_email":"aaaaaaaaa@aaaaaa.ca",
	"session_id":"temp_1776880271_73f5d1fd-f16f-4aeb-b7d1-4a703d191154"
}
*/

type GenericPaymentRequest struct {
	PayerName    string `json:"payer_name" binding:"required"`
	PayerSurname string `json:"payer_surname" binding:"required"`
	PayerEmail   string `json:"payer_email" binding:"required"`
	PayerPhone   string `json:"payer_phone" binding:"required"`
	SessionID    string `json:"session_id" binding:"required"`
}
