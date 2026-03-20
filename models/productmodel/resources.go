package productmodel

import (
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/util"
)

type SpecificationResource struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ProductResource struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Stock    int             `json:"stock"`
	Price    decimal.Decimal `json:"price"`
	Discount decimal.Decimal `json:"discount"`
	Featured bool            `json:"featured"`
	Code     string          `json:"code"`
	UUID     string          `json:"uuid"`

	// Earnings modifiers
	PlatformCut     decimal.Decimal `json:"platform_cut"`
	PlatformCutType string          `json:"platform_cut_type"`

	Colors []string `json:"colors,omitempty"`
	Sizes  []string `json:"sizes,omitempty"`

	ModerationStatus string `json:"moderation_status"`

	// Specifications
	Category            string                  `json:"category"`
	ImageURL            string                  `json:"imageUrl"`
	Description         string                  `json:"description"`
	SpecificationImages []string                `json:"specificationsImages"`
	DescriptionImages   []string                `json:"descriptionImages"`
	DeliveryImages      []string                `json:"deliveryImages"`
	Specifications      []SpecificationResource `json:"specifications"`
}

func (p Product) ToResource() ProductResource {

	if p.PlatformCut == nil {
		p.PlatformCut = &EarningsModifier{}
	}

	return ProductResource{
		ID:                  p.ID,
		Name:                p.Name,
		Stock:               p.Stock,
		Price:               p.Price,
		Discount:            p.Discount,
		Featured:            p.Featured,
		Description:         p.Description,
		ModerationStatus:    p.ModerationStatus,
		Colors:              p.Attributes.Colors,
		Sizes:               p.Attributes.Sizes,
		PlatformCut:         p.PlatformCut.Value,
		PlatformCutType:     p.PlatformCut.Type,
		Category:            p.Category,
		ImageURL:            p.Image,
		SpecificationImages: p.SpecImages,
		DescriptionImages:   p.DescImages,
		DeliveryImages:      p.PackImages,
		Specifications:      util.ResourceCollection(p.Specifications),
		Code:                p.Code,
		UUID:                p.UUID,
	}
}

func (s Specification) ToResource() SpecificationResource {
	return SpecificationResource{
		Name:  s.Name,
		Value: s.Value,
	}
}
