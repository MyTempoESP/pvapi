package imagemodel

import (
	"context"
	"strings"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
)

type imageURL string

type Image struct {
	Filename imageURL `json:"filename"`
	Alt      string   `json:"alt"`
	Tag      string   `json:"tag"`
	Product  int      `json:"product_id"`
}

func NewUrl(s string) imageURL {
	imageURL := imageURL(strings.Replace(
		s,
		".",
		"https://pvauto.linhazero.com/", 1,
	))

	return imageURL
}

type ImageCollection []Image

func NewImage(imageURL imageURL, tag string, product int) Image {
	return Image{
		Filename: imageURL,
		Alt:      slug.Make(string(imageURL)),
		Tag:      tag,
		Product:  product,
	}
}

func NewTaggedImageCollection(
	imageURLs []string,
	tag string,
	product int,
) (imc ImageCollection) {
	for _, imageURL := range imageURLs {
		imc = append(imc, NewImage(NewUrl(imageURL), tag, product))
	}

	return
}

func (imc ImageCollection) Queue(batch *pgx.Batch) (err error) {
	for _, im := range imc {
		batch.Queue(
			query.CreateImageQuery,
			im.Filename,
			im.Alt,
			im.Tag,
			im.Product,
		)
	}

	return
}

func (imc ImageCollection) Batch() (batch *pgx.Batch, err error) {
	imc.Queue(batch)
	return
}

func (im Image) Create(
	ctx context.Context,
	q db.Queryable,
) (err error) {
	_, err = q.Exec(
		ctx,
		query.CreateImageQuery,
		im.Filename,
		im.Alt,
		im.Tag,
		im.Product,
	)

	return
}
