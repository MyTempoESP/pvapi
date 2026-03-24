package shopmodel

import "linhazero.com/pvapi/address"

type CreateShopRequest struct {
	Name      string                  `json:"name" binding:"required"`
	Telephone string                  `json:"telephone" binding:"required"`
	Status    string                  `json:"status" binding:"required"`
	Email     string                  `json:"email" binding:"required"`
	Address   *address.AddressRequest `json:"address" binding:"required"`
}

type UpdateShopRequest struct {
	Name      *string                 `json:"name"`
	Telephone *string                 `json:"telephone"`
	Status    *string                 `json:"status"`
	Email     *string                 `json:"email"`
	Address   *address.AddressRequest `json:"address"`
}
