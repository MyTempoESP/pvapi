package consignmentmodel

import (
	"context"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type Consignment struct {
	ID    int
	Field any
}

func (c Consignment) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	update := query.Update("consignments AS consignment")

	update.Set.Add("field", c.Field)

	update.Where.AddMap("consignment.id", "consignment_id", c.ID)

	query, params := update.Build()

	_, err = base.Exec(ctx, query, params)

	return err
}

func (c ConsignmentBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	return model.Create(
		ctx,
		base,
		query.CreateConsignmentQuery,
		util.Args{
			"field": c.Field,
		},
	)
}

func List(
	ctx context.Context,
	base db.Queryable,
) ([]Consignment, error) {
	return util.WithCollect[Consignment](
		ctx,
		base,
		query.ListAllConsignmentsQuery,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (Consignment, error) {
	return util.WithCollectExactlyOneRow[Consignment](
		ctx,
		base,
		query.FindConsignmentQuery,
		id,
	)
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteConsignmentQuery, id)
	return err
}
