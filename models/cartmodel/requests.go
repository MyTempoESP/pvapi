package cartmodel

/*
personalization	{}
product_id	5
quantity	1
session_id	"temp_69af0ba7865c0_1773079463"
*/
type AddToCartRequest struct {
	ProductID       int            `json:"product_id" binding:"required,min=1"`
	Quantity        int            `json:"quantity" binding:"required,min=1"`
	Personalization map[string]any `json:"personalization"`
	SessionID       string         `json:"session_id" binding:"required"`
}
