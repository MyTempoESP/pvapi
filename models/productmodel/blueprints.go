package productmodel

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"
	"github.com/shopspring/decimal"
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

type ProductBlueprint struct {
	Name     string
	Price    string
	Stock    int
	Discount decimal.Decimal
	Category string
	Featured bool

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

	uniqueid, err := uuid.NewV4()

	if err != nil {
		return
	}

	categorySlug := slug.Make(p.Category)

	// Generate the default specifications AFTER user-definitions
	specifications = append(specifications,
		SpecificationBlueprint{
			Name:  "category",
			Value: categorySlug,
		},
		SpecificationBlueprint{
			Name:  "description",
			Value: p.Description,
		},
		SpecificationBlueprint{
			Name:  "uuid",
			Value: uniqueid.String(),
		},
	)

	code := slug.Make(p.Name) + categorySlug

	// at least of the form: X, description, category
	if len(specifications) > 2 {
		code += specifications[0].Value
	}

	specifications = append(specifications, SpecificationBlueprint{
		Name:  "code",
		Value: code,
	})

	return
}

func PreparePolicies(
	p *CreateProductRequest,
) (policies []PolicyBlueprint) {
	policies = append(policies,
		PolicyBlueprint{
			Name:    "colors",
			Value:   p.ColorListJSON,
			Enabled: true,
		},
		PolicyBlueprint{
			Name:    "customName",
			Value:   "{}",
			Enabled: p.CustomName,
		},
		PolicyBlueprint{
			Name:    "customModality",
			Value:   "{}",
			Enabled: p.customModality,
		},
		PolicyBlueprint{
			Name:    "customColorSelection",
			Value:   p.ColorListJSON,
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

	pb = &ProductBlueprint{
		Name:     p.Name,
		Price:    p.Price,
		Stock:    p.Stock,
		Discount: p.Discount,
		Featured: p.Featured,

		Specifications: specifications,
		Policies:       policies,

		ImageURL:   image,
		DescImages: descImages,
		SpecImages: specImages,
		PackImages: packImages,
	}

	return
}
