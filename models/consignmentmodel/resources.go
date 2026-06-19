package consignmentmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

type ProductResource struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	Code             string          `json:"code"`
	UUID             string          `json:"uuid"`
	Image            string          `json:"image"`
	ImageURL         string          `json:"imageUrl"`
	Category         string          `json:"category"`
	CategoryID       int             `json:"category_id"`
	Description      string          `json:"description"`
	OriginalPrice    decimal.Decimal `json:"original_price"`
	OriginalQuantity int             `json:"original_quantity"`
	Price            decimal.Decimal `json:"price"`
	Quantity         int             `json:"quantity"`
	Status           string          `json:"status"`
	Discount         decimal.Decimal `json:"discount"`
	DiscountType     string          `json:"discount_type"`

	// ponytail: duplicated for backwards compatibility
	ConsignmentPrice    decimal.Decimal `json:"consignment_price"`
	ConsignmentQuantity int             `json:"consignment_quantity"`
	Featured            bool            `json:"featured"`

	Colors []string `json:"colors,omitempty"`
	Sizes  []string `json:"sizes,omitempty"`

	AllowCustomName           bool `json:"allowCustomName"`
	AllowCustomModality       bool `json:"allowCustomModality"`
	AllowCustomColorSelection bool `json:"allowCustomColorSelection"`

	PlatformCut     decimal.Decimal `json:"platform_cut"`
	PlatformCutType string          `json:"platform_cut_type"`

	Specifications      []productSpecResource `json:"specifications"`
	SpecificationImages []string              `json:"specificationsImages"`
	DescriptionImages   []string              `json:"descriptionImages"`
	DeliveryImages      []string              `json:"deliveryImages"`
}

type productSpecResource struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DestinationsResource []PickupResource

type PickupResource struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewPickupResourceFromPickup(p Pickup) PickupResource {
	return PickupResource(p)
}

type ConsignmentResource struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	ComissionType  string          `json:"commission_type"`
	ComissionValue decimal.Decimal `json:"commission_value"`
	Code           string          `json:"code"`
	Slug           string          `json:"slug"`
	Status         string          `json:"status"`

	Destinations  DestinationsResource `json:"destinations"`
	ConsigneeName string               `json:"responsibleMonitor"`

	CreatedAt time.Time `json:"created_at"`
}

type ConsignmentWithStatsResource struct {
	Code              string               `json:"code"`
	ID                int                  `json:"id"`
	Name              string               `json:"name"`
	Status            string               `json:"status"`
	Destinations      DestinationsResource `json:"destinations"`
	MonitorName       string               `json:"monitorName"`
	CreatedAt         time.Time            `json:"createdAt"`
	TotalProducts     int                  `json:"totalProducts"`
	ConfirmedProducts int                  `json:"confirmedProducts"`
}

func (c ConsignmentWithStats) ToResource() ConsignmentWithStatsResource {
	return ConsignmentWithStatsResource{
		Code:              c.Code,
		ID:                c.ID,
		Name:              c.Name,
		Status:            c.Status,
		Destinations:      pickupsToDestinations(c.Pickups),
		MonitorName:       c.Consignee.Name,
		CreatedAt:         c.CreatedAt,
		TotalProducts:     c.TotalProducts,
		ConfirmedProducts: c.ConfirmedProducts,
	}
}

type PublicConsignmentResource struct {
	ID             int             `json:"id"`
	Name           string          `json:"name"`
	ComissionType  string          `json:"commission_type"`
	ComissionValue decimal.Decimal `json:"commission_value"`
	Code           string          `json:"code"`
	Slug           string          `json:"slug"`
	Status         string          `json:"status"`

	Destinations  DestinationsResource `json:"destinations"`
	ConsigneeName string               `json:"responsibleMonitor"`

	Shop struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"shop"`

	CreatedAt time.Time `json:"created_at"`
}

func pickupsToDestinations(pickups []Pickup) DestinationsResource {
	if pickups == nil {
		return DestinationsResource{}
	}
	res := make(DestinationsResource, len(pickups))
	for i, p := range pickups {
		res[i] = NewPickupResourceFromPickup(p)
	}
	return res
}

func (c Consignment) ToResource() ConsignmentResource {
	return ConsignmentResource{
		ID:             c.ID,
		Name:           c.Name,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		Code:           c.Code,
		Slug:           c.Slug,
		Status:         c.Status,

		Destinations:  pickupsToDestinations(c.Pickups),
		ConsigneeName: c.Consignee.Name,

		CreatedAt: c.CreatedAt,
	}
}

func (c Consignment) ToPublicResource() PublicConsignmentResource {
	r := PublicConsignmentResource{
		ID:             c.ID,
		Name:           c.Name,
		ComissionType:  c.ComissionType,
		ComissionValue: c.ComissionValue,
		Code:           c.Code,
		Slug:           c.Slug,
		Status:         c.Status,

		Destinations:  pickupsToDestinations(c.Pickups),
		ConsigneeName: c.Consignee.Name,

		CreatedAt: c.CreatedAt,
	}
	r.Shop.ID = c.Shop.ID
	r.Shop.Name = c.Shop.Name
	return r
}

func (p ConsignmentProduct) ToResource() ProductResource {
	specs := make([]productSpecResource, len(p.Attributes.Specifications))
	for i, s := range p.Attributes.Specifications {
		specs[i] = productSpecResource{Name: s.Name, Value: s.Value}
	}

	return ProductResource{
		ID:               p.ID,
		Name:             p.Name,
		Code:             p.Code,
		UUID:             p.UUID,
		Image:            p.Attributes.Image,
		ImageURL:         p.Attributes.Image,
		Category:         p.Category,
		CategoryID:       p.CategoryID,
		Description:      p.Description,
		OriginalPrice:    p.Product.Price,
		OriginalQuantity: p.Stock,
		Price:            p.Package.Price,
		Quantity:         p.Quantity,
		Status:           p.Status,
		Discount:         p.Package.Discount,
		DiscountType:     p.DiscountType,

		ConsignmentPrice:    p.Package.Price,
		ConsignmentQuantity: p.Quantity,
		Featured:            p.Featured,

		Colors: p.Attributes.Colors,
		Sizes:  p.Attributes.Sizes,

		AllowCustomName:           p.Attributes.AllowCustomName,
		AllowCustomModality:       p.Attributes.AllowCustomModality,
		AllowCustomColorSelection: p.Attributes.AllowCustomColor,

		PlatformCut:     p.PlatformCut,
		PlatformCutType: p.PlatformCutType,

		Specifications:      specs,
		SpecificationImages: p.Attributes.SpecImages,
		DescriptionImages:   p.Attributes.DescImages,
		DeliveryImages:      p.Attributes.PackImages,
	}
}
