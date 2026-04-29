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
	PayerName    string `json:"payer_name"`
	PayerSurname string `json:"payer_surname"`
	PayerEmail   string `json:"payer_email"`
	SessionID    string `json:"session_id"`
}

