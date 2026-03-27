package productmodel

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/util"
)

type SpecificationBlueprint struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PolicyBlueprint struct {
	Name    string
	Value   string
	Enabled bool
}

type AttributesBlueprint struct {
	Colors []string `json:"colors,omitempty"`
	Sizes  []string `json:"sizes,omitempty"`
}

type ProductBlueprint struct {
	Name        string
	Price       string
	Stock       int
	Discount    decimal.Decimal
	Category    string
	Featured    bool
	Code        string
	Description string
	Attributes  []byte

	ImageURL   string
	DescImages []string
	SpecImages []string
	PackImages []string

	Specifications []SpecificationBlueprint

	Policies []PolicyBlueprint
}

func PrepareSpecifications(
	p *CreateProductRequest,
) (specifications []SpecificationBlueprint, err error) {
	// Fetch the user specifications
	err = json.Unmarshal([]byte(p.SpecificationsJSON), &specifications)

	if err != nil {
		return nil, err
	}

	return
}

func PreparePolicies(
	p *CreateProductRequest,
) (policies []PolicyBlueprint) {
	policies = append(policies,
		PolicyBlueprint{
			Name:    "customName",
			Enabled: p.CustomName,
		},
		PolicyBlueprint{
			Name:    "customModality",
			Enabled: p.customModality,
		},
		PolicyBlueprint{
			Name:    "customColor",
			Enabled: p.CustomColorSelection,
		},
	)
	return
}

func (p *CreateProductRequest) Prepare(
	c *gin.Context,
) (pb *ProductBlueprint, err error) {

	policies := PreparePolicies(p)

	specifications, err := PrepareSpecifications(p)

	if err != nil {
		return
	}

	var colors []string
	json.Unmarshal([]byte(p.ColorListJSON), &colors)

	image, err := handleImage(c, p.Image)

	if err != nil {
		return
	}

	descImages, err := handleImages(c, p.DescImages)

	if err != nil {
		return
	}

	specImages, err := handleImages(c, p.SpecImages)

	if err != nil {
		return
	}

	packImages, err := handleImages(c, p.PackImages)

	if err != nil {
		return
	}

	code := slug.Make(p.Name) + slug.Make(p.Category)

	// at least of the form: X, description, category
	if len(specifications) > 2 {
		code += specifications[0].Value
	}

	attributes, err := json.Marshal(AttributesBlueprint{
		Colors: colors,
		Sizes:  []string{},
	})

	if err != nil {
		return
	}

	// i just realized sometimes the front-end doesnt send this.
	// great job devs.
	featured := util.Coalesce(p.Featured, false)

	pb = &ProductBlueprint{
		Name:        p.Name,
		Price:       p.Price,
		Stock:       p.Stock,
		Discount:    p.Discount,
		Featured:    featured,
		Code:        code,
		Description: p.Description,
		Attributes:  attributes,

		Specifications: specifications,
		Policies:       policies,

		ImageURL:   image,
		DescImages: descImages,
		SpecImages: specImages,
		PackImages: packImages,
	}

	return
}
