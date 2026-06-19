package consignmentmodel

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"

	"linhazero.com/pvapi/models/productmodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

func (a AddProductAction) Exec(
	ctx context.Context,
	tx pgx.Tx,
) (err error) {
	for _, p := range a.Products {
		if p.Discount.LessThan(decimal.NewFromInt(0)) || p.Discount.GreaterThan(decimal.NewFromInt(100)) {
			return fmt.Errorf(
				"discount must be between 0 and 100, got %s: %w",
				p.Discount.String(),
				util.ErrBadRequest,
			)
		}

		var exists int
		err = tx.QueryRow(
			ctx, query.ProductCheckExistsInConsignmentQuery,
			p.ID, a.ConsignmentID,
		).Scan(&exists)

		if err == nil {
			return fmt.Errorf(
				"product %d already exists in consignment %d: %w",
				p.ID, a.ConsignmentID,
				util.ErrDuplicatedResource,
			)
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(
				"could not check product %d in consignment %d: %w",
				p.ID, a.ConsignmentID, err,
			)
		}
	}

	batch := &pgx.Batch{}

	for _, p := range a.Products {
		product := p.Prepare()

		batch.Queue(
			query.ProductAddToConsignmentQuery,
			util.Args{
				"product_id":     product.ID,
				"consignment_id": a.ConsignmentID,
				"quantity":       product.Quantity,
				"price":          product.Price,
				"attr":           product.AttrVal,
				"discount":       product.Discount,
				"featured":       product.Featured,
			},
		)
	}

	err = util.Bake(ctx, tx, batch)

	if err != nil {
		return fmt.Errorf(
			"couldn't add products to consignment %q: %w",
			a.ConsignmentID,
			err,
		)
	}

	return
}

func (p UpdateConsignmentProductBlueprint) Exec(
	ctx context.Context,
	tx pgx.Tx,
) (err error) {
	q := util.SQL

	fields := make(map[string]any)

	if p.Price != nil {
		fields["price"] = *p.Price
	}

	if p.Quantity != nil {
		if err = productmodel.TakeFromPackage(
			ctx, tx,
			p.ProductID,
			p.ConsignmentID,
			*p.Quantity,
		); err != nil {
			return
		}
		fields["quantity"] = *p.Quantity
	}

	if p.Status != nil {
		fields["status"] = *p.Status
	}

	if len(fields) == 0 {
		return
	}

	query, params, err := q.Update("packages").
		SetMap(fields).
		Where(sq.And{
			sq.Eq{"product_id": p.ProductID},
			sq.Eq{"consignment_id": p.ConsignmentID},
		}).
		ToSql()

	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	_, err = tx.Exec(ctx, query, params...)

	if err != nil {
		return fmt.Errorf(
			"could not update consignment, query: %q, %w",
			query, err)
	}

	return err
}

func (c ConsignmentUpdateBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	q := util.SQL

	fields := make(map[string]any)

	if c.ConsigneeID != nil {
		fields["consignee_id"] = *c.ConsigneeID
	}

	if c.ComissionType != nil {
		fields["comission_type"] = *c.ComissionType
	}

	if c.ComissionValue != nil {
		fields["comission_value"] = *c.ComissionValue
	}

	if c.Status != nil {
		fields["status"] = *c.Status
	}

	if c.Name != nil {
		fields["name"] = *c.Name
	}

	if len(fields) == 0 {
		return
	}

	query, params, err := q.Update("consignments").
		SetMap(fields).
		Where(sq.Eq{"id": c.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	_, err = base.Exec(ctx, query, params...)

	if err != nil {
		return fmt.Errorf(
			"could not update consignment, query: %q, %w", query, err)
	}

	return err
}

func (c ConsignmentBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	id, err := model.Create(
		ctx,
		base,
		query.CreateConsignmentQuery,
		util.Args{
			"shop_id":            c.ShopID,
			"consignee_id":       c.ConsigneeID,
			"comission_type":     c.ComissionType,
			"comission_value":    c.ComissionValue,
			"status":             c.Status,
			"name":               c.Name,
		},
	)
	if err != nil {
		return 0, err
	}

	if len(c.PickupIDs) > 0 {
		for _, pickupID := range c.PickupIDs {
			_, err = base.Exec(
				ctx,
				query.AddConsignmentPickupQuery,
				id, pickupID,
			)
			if err != nil {
				return 0, fmt.Errorf(
					"could not add pickup %d to consignment %d: %w",
					pickupID, id, err,
				)
			}
		}
	}

	return id, nil
}

type pickupRow struct {
	ConsignmentID int    `db:"consignment_id"`
	PickupID      int    `db:"pickup_location_id"`
	PickupName    string `db:"pickup_location_name"`
}

func BatchFetchConsignmentPickups(
	ctx context.Context,
	base db.Queryable,
	consignmentIDs []int,
) (map[int][]Pickup, error) {
	if len(consignmentIDs) == 0 {
		return map[int][]Pickup{}, nil
	}

	ids := make([]any, len(consignmentIDs))
	for i, id := range consignmentIDs {
		ids[i] = id
	}

	rows, err := util.WithCollect[pickupRow](
		ctx, base,
		query.BatchFetchConsignmentPickupsQuery,
		ids,
	)
	if err != nil {
		return nil, err
	}

	result := make(map[int][]Pickup, len(consignmentIDs))
	for _, r := range rows {
		result[r.ConsignmentID] = append(result[r.ConsignmentID], Pickup{
			ID:   r.PickupID,
			Name: r.PickupName,
		})
	}

	for _, id := range consignmentIDs {
		if _, ok := result[id]; !ok {
			result[id] = []Pickup{}
		}
	}

	return result, nil
}

func (a AddPickupAction) Exec(
	ctx context.Context,
	tx pgx.Tx,
) error {
	_, err := tx.Exec(
		ctx, query.AddConsignmentPickupQuery,
		a.ConsignmentID, a.PickupID,
	)
	if err != nil {
		return fmt.Errorf(
			"could not add pickup %d to consignment %d: %w",
			a.PickupID, a.ConsignmentID, err,
		)
	}
	return nil
}

func (a RemovePickupAction) Exec(
	ctx context.Context,
	tx pgx.Tx,
) error {
	_, err := tx.Exec(
		ctx, query.RemoveConsignmentPickupQuery,
		a.ConsignmentID, a.PickupID,
	)
	if err != nil {
		return fmt.Errorf(
			"could not remove pickup %d from consignment %d: %w",
			a.PickupID, a.ConsignmentID, err,
		)
	}
	return nil
}

type CodeLookup struct {
	ID     int    `db:"id"`
	Status string `db:"status"`
}

func FindByCode(
	ctx context.Context,
	base db.Queryable,
	code string,
) (CodeLookup, error) {
	return util.WithCollectExactlyOneRow[CodeLookup](
		ctx, base,
		query.FindConsignmentByCodeQuery,
		code,
	)
}

func List(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]Consignment, error) {
	consignments, err := util.WithCollect[Consignment](
		ctx,
		base,
		query.ListConsignmentsByShopQuery,
		shopID,
	)
	if err != nil {
		return nil, err
	}

	return enrichWithPickups(ctx, base, consignments)
}

func ListAllByShop(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]Consignment, error) {
	consignments, err := util.WithCollect[Consignment](
		ctx,
		base,
		query.ListAllConsignmentsByShopQuery,
		shopID,
	)
	if err != nil {
		return nil, err
	}

	return enrichWithPickups(ctx, base, consignments)
}

func ListByConsignee(
	ctx context.Context,
	base db.Queryable,
	consigneeID int,
) ([]Consignment, error) {
	consignments, err := util.WithCollect[Consignment](
		ctx,
		base,
		query.ListConsignmentsByConsigneeQuery,
		consigneeID,
	)
	if err != nil {
		return nil, err
	}

	return enrichWithPickups(ctx, base, consignments)
}

func ListByConsigneeWithStats(
	ctx context.Context,
	base db.Queryable,
	consigneeID int,
) ([]ConsignmentWithStats, error) {
	consignments, err := util.WithCollect[ConsignmentWithStats](
		ctx,
		base,
		query.ListConsignmentsByConsigneeQuery,
		consigneeID,
	)
	if err != nil {
		return nil, err
	}

	return enrichWithStatsPickups(ctx, base, consignments)
}

func ListAllByConsigneeWithStats(
	ctx context.Context,
	base db.Queryable,
	consigneeID int,
) ([]ConsignmentWithStats, error) {
	consignments, err := util.WithCollect[ConsignmentWithStats](
		ctx,
		base,
		query.ListAllConsignmentsByConsigneeQuery,
		consigneeID,
	)
	if err != nil {
		return nil, err
	}

	return enrichWithStatsPickups(ctx, base, consignments)
}

func PublicList(
	ctx context.Context,
	base db.Queryable,
) ([]Consignment, error) {
	consignments, err := util.WithCollect[Consignment](
		ctx,
		base,
		query.ListAllActiveConsignmentsQuery,
	)
	if err != nil {
		return nil, err
	}

	return enrichWithPickups(ctx, base, consignments)
}

func PublicListFiltered(
	ctx context.Context,
	base db.Queryable,
	shopIDs []int,
) ([]Consignment, error) {
	if len(shopIDs) == 0 {
		return PublicList(ctx, base)
	}

	ids := make([]any, len(shopIDs))
	for i, id := range shopIDs {
		ids[i] = id
	}

	q := util.SQL.Select(
		"consignment_id",
		"consignment_name",
		"consignment_comission_type",
		"consignment_comission_value",
		"consignment_status",
		"consignment_created_at",
		"consignment_slug",
		"consignment_code",

		"consignee_id",
		"consignee_name",

		"shop_id",
		"shop_name",
	).From("consignments_listing_view").
		Where("consignment_id IN (SELECT id FROM consignments_active)").
		Where(sq.Eq{"shop_id": shopIDs})

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build shop-filter consignment query: %w", err)
	}

	consignments, err := util.WithCollect[Consignment](ctx, base, sql, args...)
	if err != nil {
		return nil, err
	}

	return enrichWithPickups(ctx, base, consignments)
}

func ListProducts(
	ctx context.Context,
	base db.Queryable,
	id any,
) ([]ConsignmentProduct, error) {
	return util.WithCollect[ConsignmentProduct](
		ctx,
		base,
		query.ProductListInConsignmentQuery,
		id,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (DetailedConsignment, error) {
	consignment, err := util.WithCollectExactlyOneRow[DetailedConsignment](
		ctx,
		base,
		query.FindConsignmentQuery,
		id,
	)
	if err != nil {
		return consignment, err
	}

	intID, ok := id.(int)
	if !ok {
		return consignment, nil
	}

	pickupMap, err := BatchFetchConsignmentPickups(
		ctx, base, []int{intID},
	)
	if err != nil {
		return consignment, err
	}

	consignment.Pickups = pickupMap[intID]
	return consignment, nil
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteConsignmentQuery, id)
	return err
}

func enrichWithPickups(
	ctx context.Context,
	base db.Queryable,
	consignments []Consignment,
) ([]Consignment, error) {
	if len(consignments) == 0 {
		return consignments, nil
	}

	ids := make([]int, len(consignments))
	for i, c := range consignments {
		ids[i] = c.ID
	}

	pickupMap, err := BatchFetchConsignmentPickups(ctx, base, ids)
	if err != nil {
		return nil, err
	}

	for i := range consignments {
		consignments[i].Pickups = pickupMap[consignments[i].ID]
	}

	return consignments, nil
}

func enrichWithStatsPickups(
	ctx context.Context,
	base db.Queryable,
	consignments []ConsignmentWithStats,
) ([]ConsignmentWithStats, error) {
	if len(consignments) == 0 {
		return consignments, nil
	}

	ids := make([]int, len(consignments))
	for i, c := range consignments {
		ids[i] = c.ID
	}

	pickupMap, err := BatchFetchConsignmentPickups(ctx, base, ids)
	if err != nil {
		return nil, err
	}

	for i := range consignments {
		consignments[i].Pickups = pickupMap[consignments[i].ID]
	}

	return consignments, nil
}
