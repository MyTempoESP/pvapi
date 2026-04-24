package cartmodel

/*
personalization	{}
product_id	5
quantity	1
session_id	"temp_69af0ba7865c0_1773079463"
*/
type AddToCartRequest struct {
	ProductID       int            `json:"product_id"`
	Quantity        int            `json:"quantity"`
	Personalization map[string]any `json:"personalization"`
	SessionID       string         `json:"session_id"`
}
