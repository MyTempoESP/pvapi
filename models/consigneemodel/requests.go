package consigneemodel

import "linhazero.com/pvapi/address"

type CreateConsigneeRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required"`
	Phone   string `json:"phone" binding:"required"`
	Address *address.AddressRequest
}

type UpdateConsigneeRequest struct {
	Name    *string                       `json:"name"`
	Email   *string                       `json:"email"`
	Phone   *string                       `json:"phone"`
	Address *address.AddressUpdateRequest `json:"address"`
}
