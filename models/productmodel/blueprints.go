package productmodel

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/util"
)

type ProductBlueprint struct {
	Shop        int
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
}

func (p *CreateProductRequest) Prepare(
	c *gin.Context,
	shop int,
) (pb *ProductBlueprint, err error) {

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

	var a ProductAttr

	err = json.Unmarshal([]byte(p.SpecificationsJSON), &a.Specifications)

	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(p.ColorListJSON), &a.Colors)

	if err != nil {
		return
	}

	a.AllowCustomColor = p.CustomColorSelection
	a.AllowCustomModality = p.CustomModality
	a.AllowCustomName = p.CustomName

	attributes, err := json.Marshal(a)

	if err != nil {
		return
	}

	code := slug.Make(p.Name) + slug.Make(p.Category)

	// at least of the form: X, description, category
	if len(a.Specifications) > 2 {
		code += a.Specifications[0].Value
	}

	// i just realized sometimes the front-end doesnt send this.
	// great job devs.
	featured := util.Coalesce(p.Featured, false)

	pb = &ProductBlueprint{
		Shop:        shop,
		Name:        p.Name,
		Price:       p.Price,
		Stock:       p.Stock,
		Discount:    p.Discount,
		Featured:    featured,
		Code:        code,
		Description: p.Description,
		Attributes:  attributes,

		ImageURL:   image,
		DescImages: descImages,
		SpecImages: specImages,
		PackImages: packImages,
	}

	return
}
