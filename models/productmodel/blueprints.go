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
	CategoryID  int
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

	logFailedUpload := func(c *gin.Context, imgType string, err error) {
		util.LogWarn(
			c, "tried (and failed) to upload image",
			"shop", shop,
			"error", err,
			"imgType", imgType,
			"product", product,
		)
	}

	if p.Image != nil {
		image, err := handleImage(c, p.Image)

		if err == nil {
			pb.ImageURL = &image
		} else {
			logFailedUpload(c, "main", err)
		}
	}

	if p.DescImages != nil {
		images, err := handleImages(c, p.DescImages)

		if err == nil {
			pb.DescImagesURLs = images
		} else {
			logFailedUpload(c, "description", err)
		}
	}

	if p.SpecImages != nil {
		images, err := handleImages(c, p.SpecImages)

		if err == nil {
			pb.SpecImagesURLs = images
		} else {
			logFailedUpload(c, "specifications", err)
		}
	}

	if p.PackImages != nil {
		images, err := handleImages(c, p.PackImages)

		if err == nil {
			pb.PackImagesURLs = images
		} else {
			logFailedUpload(c, "delivery", err)
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

	code := slug.Make(p.Name)

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
		CategoryID:  p.CategoryID,
		Discount:    p.Discount,
		Featured:    featured,
		Code:        code,
		Description: p.Description,
		Attributes:  attributes,
	}

	return
}
