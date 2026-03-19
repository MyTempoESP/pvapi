package imagemodel

import (
	"context"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
)

type Image struct {
	Filename URL
	Alt      string
	Tag      string
	Product  int
}

type URL string

type ImageCollection []Image

func NewImage(url URL, tag string, product int) Image {
	return Image{
		Filename: url,
		Alt:      slug.Make(string(url)),
		Tag:      tag,
		Product:  product,
	}
}

func NewTaggedImageCollection(
	urls []string,
	tag string,
	product int,
) (imc ImageCollection) {
	for _, url := range urls {
		imc = append(imc, NewImage(URL(url), tag, product))
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
