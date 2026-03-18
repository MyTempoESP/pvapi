package productmodel

import "github.com/shopspring/decimal"

type SpecificationResource struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ProductResource struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Stock    int             `json:"code"`
	Price    decimal.Decimal `json:"price"`
	Discount decimal.Decimal `json:"discount"`
	Featured bool            `json:"featured"`

	// Earnings modifiers
	PlatformCut     decimal.Decimal `json:"platform_cut"`
	PlatformCutType string          `json:"platform_cut_type"`

	// Product policies
	Colors           []string `json:"colors"`
	ModerationStatus bool     `json:"moderation_status"`

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

	return ProductResource{
		ID:       p.ID,
		Name:     p.Name,
		Stock:    p.Stock,
		Price:    p.Price,
		Discount: p.Discount,
		Featured: p.Featured,
	}
}
