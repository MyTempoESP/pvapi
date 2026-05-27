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

	AllowCustomName           bool `json:"allowCustomName"`
	AllowCustomModality       bool `json:"allowCustomModality"`
	AllowCustomColorSelection bool `json:"allowCustomColorSelection"`

	ModerationStatus string `json:"moderation_status"`

	// Specifications
	Category            string                  `json:"category"`
	CategoryID          int                     `json:"category_id"`
	ImageURL            string                  `json:"imageUrl"`
	Description         string                  `json:"description"`
	SpecificationImages []string                `json:"specificationsImages"`
	DescriptionImages   []string                `json:"descriptionImages"`
	DeliveryImages      []string                `json:"deliveryImages"`
	Specifications      []SpecificationResource `json:"specifications"`

	Consignments []int `json:"consignments,omitempty"`
}

/*
'id' => (string) $this->id,
'name' => (string) $this->name,
'description' => (string) $this->description,
'price' => (float) $this->price,
'category' => $this->category ? $this->category->name : 'Sem categoria',
'imageUrl' => $this->image ? url('storage/' . $this->image) : null,
'establishmentId' => (string) $this->shop_id,
'establishmentName' => $this->shop ? $this->shop->name : 'Sem loja',
'taxType' => $this->platform_cut_type === 'fixed' ? 'fixed' : 'percentage',
*/
type PendingProductResource struct {
	ID              int             `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Price           decimal.Decimal `json:"price"`
	Category        string          `json:"category"`
	ImageURL        string          `json:"imageUrl"`
	ShopID          int             `json:"establishmentId"`
	ShopName        string          `json:"establishmentName"`
	PlatformCutType string          `json:"taxType"`
}

func (p Product) ToPendingResource() PendingProductResource {
	return PendingProductResource{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		Price:           p.Price,
		Category:        p.Category,
		ImageURL:        p.Attributes.Image,
		ShopID:          p.ShopID,
		ShopName:        p.ShopName,
		PlatformCutType: p.PlatformCutType,
	}
}

func (p Product) ToResource() ProductResource {

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
		PlatformCut:         p.PlatformCut,
		PlatformCutType:     p.PlatformCutType,
		Category:            p.Category,
		CategoryID:          p.CategoryID,
		ImageURL:            p.Attributes.Image,
		SpecificationImages: p.Attributes.SpecImages,
		DescriptionImages:   p.Attributes.DescImages,
		DeliveryImages:      p.Attributes.PackImages,

		AllowCustomColorSelection: p.Attributes.AllowCustomColor,
		AllowCustomModality:       p.Attributes.AllowCustomModality,
		AllowCustomName:           p.Attributes.AllowCustomName,

		Specifications: util.ResourceCollection(
			p.Attributes.Specifications,
		),

		Code: p.Code,
		UUID: p.UUID,

		Consignments: p.Consignments,
	}
}

func (s Specification) ToResource() SpecificationResource {
	return SpecificationResource(s)
}
