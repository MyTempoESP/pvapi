package consigneemodel

import "linhazero.com/pvapi/address"

type ConsigneeResource struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Email   string          `json:"email"`
	Phone   string          `json:"phone"`
	Address address.Address `json:"address"`
}

func (c Consignee) ToResource() *ConsigneeResource {
	return &ConsigneeResource{
		ID:      c.ID,
		Name:    c.User.Name,
		Email:   c.User.Email,
		Phone:   c.Phone,
		Address: c.Address,
	}
}
