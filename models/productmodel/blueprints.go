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
	Description string
	Code        string
	Attributes  []byte
}

type UpdateProductBlueprint struct {
	ID   int
	Shop int

	ImageURL       *string
	DescImagesURLs []string
	SpecImagesURLs []string
	PackImagesURLs []string

	UpdateProductRequest
}

func (p *UpdateProductRequest) Prepare(
	c *gin.Context,
	product, shop int,
) (pb *UpdateProductBlueprint, err error) {
	pb = &UpdateProductBlueprint{
		ID:                   product,
		Shop:                 shop,
		UpdateProductRequest: *p,
	}

	util.LogInfo(c, "trying to handle images")

	if p.Image != nil {
		var image string
		image, err = handleImage(c, p.Image)

		if err != nil {
			return
		}

		pb.ImageURL = &image
	}

	if p.DescImages != nil {
		pb.DescImagesURLs, err = handleImages(c, p.DescImages)

		if err != nil {
			return
		}
	}

	if p.SpecImages != nil {
		pb.SpecImagesURLs, err = handleImages(c, p.SpecImages)

		if err != nil {
			return
		}
	}

	if p.PackImages != nil {
		pb.PackImagesURLs, err = handleImages(c, p.PackImages)

		if err != nil {
			return
		}
	}

	return
}

func (p *CreateProductRequest) Prepare(
	c *gin.Context,
	shop int,
) (pb *ProductBlueprint, err error) {
	var a ProductAttr

	a.Image, err = handleImage(c, p.Image)

	if err != nil {
		return
	}

	a.DescImages, err = handleImages(c, p.DescImages)

	if err != nil {
		return
	}

	a.SpecImages, err = handleImages(c, p.SpecImages)

	if err != nil {
		return
	}

	a.PackImages, err = handleImages(c, p.PackImages)

	if err != nil {
		return
	}

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
		Category:    p.Category,
		Discount:    p.Discount,
		Featured:    featured,
		Code:        code,
		Description: p.Description,
		Attributes:  attributes,
	}

	return
}
