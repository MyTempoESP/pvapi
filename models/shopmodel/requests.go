package shopmodel

type addressField struct {
	Street       string `json:"street" binding:"required"`
	Number       string `json:"number" binding:"required"`
	Neighborhood string `json:"neighborhood" binding:"required"`
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	CEP          string `json:"CEP" binding:"required"`
	Complement   string `json:"complement" binding:"required"`
}

type CreateShopRequest struct {
	Name      string       `json:"name" binding:"required"`
	Telephone string       `json:"telephone" binding:"required"`
	Status    string       `json:"status" binding:"required"`
	Address   addressField `json:"address" binding:"required"`
}
