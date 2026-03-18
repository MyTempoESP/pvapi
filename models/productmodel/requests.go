package productmodel

import "github.com/shopspring/decimal"

type Policies struct {
	SelectColors         []string `form:"colors" binding:"required"`
	CustomName           bool     `form:"allowCustomName" binding:"required"`
	CustomModality       bool     `form:"allowCustomModality" binding:"required"`
	CustomColorSelection bool     `form:"allowCustomColorSelection" binding:"required"`
}

type SpecificationField struct {
	Name  string `form:"name" binding:"required"`
	Value string `form:"value" binding:"required"`
}

type CategorySpecification struct {
	Name  string `form:"name" binding:"required"`
	Value string `form:"slug" binding:"required"`
}

type ProductField struct {
	Name     string          `form:"name" binding:"required"`
	Price    decimal.Decimal `form:"price" binding:"required"`
	Discount decimal.Decimal `form:"discount" binding:"required"`
	Featured bool            `form:"featured" binding:"required"`

	// Specification
	Description string `form:"description" binding:"required"`
}

type CreateProductRequest struct {
	Product ProductField `form:"product" binding:"required"`

	// Specifications
	Specifications []SpecificationField `form:"specifications" binding:"required"`

	// Meta-Specifications
	Category CategorySpecification `form:"category" binding:"required"`

	// Policies
	Customizations Policies `form:"customization" binding:"required"`

	// Quantity
	Stock int `form:"quantity" binding:"required"`
}
