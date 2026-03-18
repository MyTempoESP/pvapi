package productmodel

import "github.com/shopspring/decimal"

type CreateProductRequest struct {
	Name     string          `form:"name" binding:"required"`
	Price    decimal.Decimal `form:"price" binding:"required"`
	Stock    int             `form:"stock" binding:"required"`
	Discount decimal.Decimal `form:"discount" binding:"required"`
	Featured bool            `form:"featured" binding:"required"`

	// Specifications
	Description string `form:"description" binding:"required"`
	Category    string `form:"category" binding:"required"`

	// Images

}
