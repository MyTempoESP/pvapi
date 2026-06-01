package consigneemodel

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/address"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type Consignee struct {
	ID    int    `db:"consignee_id"`
	Phone string `db:"phone"`

	// anonymous fields
	address.Address
	usermodel.User
}

func (c ConsigneeBlueprint) Save(
	ctx context.Context,
	base db.Queryable,
) (id int, err error) {
	userID, err := c.User.MustSave(ctx, base)

	if err != nil {
		return
	}

	id, err = model.Create(
		ctx,
		base,
		query.CreateConsigneeQuery,
		util.Merge(
			pgx.NamedArgs{
				"telephone": c.Phone,
				"shop_id":   c.ShopID,
				"user_id":   userID,
			},
			c.Address.Prepare().Params(),
		),
	)

	return
}

func List(
	ctx context.Context,
	pool *pgxpool.Pool,
	shop int,
) ([]Consignee, error) {
	return util.WithCollect[Consignee](
		ctx,
		pool,
		query.ListShopConsigneesQuery,
		shop,
	)
}

func HasConsignment(
	ctx context.Context,
	base db.Queryable,
	consigneeID, consignmentID int,
) (bool, error) {
	return model.HasResource(
		ctx,
		base,
		query.ConsigneeHasConsignment,
		consigneeID, consignmentID,
	)
}

func HasProduct(
	ctx context.Context,
	base db.Queryable,
	consigneeID, productID int,
) (bool, error) {
	return model.HasResource(
		ctx,
		base,
		query.ConsigneeHasProduct,
		consigneeID, productID,
	)
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	consigneeID int,
) error {
	_, err := base.Exec(ctx, query.DeleteConsigneeQuery, consigneeID)
	return err
}

func FindByID(
	ctx context.Context,
	base db.Queryable,
	consigneeID int,
) (Consignee, error) {
	return util.WithCollectExactlyOneRow[Consignee](
		ctx, base,
		query.FindConsigneeByIDQuery,
		consigneeID,
	)
}

func Update(
	ctx context.Context,
	base db.Queryable,
	consigneeID int,
	req UpdateConsigneeRequest,
) error {
	if req.Name != nil || req.Email != nil {
		_, err := base.Exec(
			ctx, query.UpdateConsigneeUserQuery,
			consigneeID, req.Name, req.Email,
		)
		if err != nil {
			return err
		}
	}

	if req.Phone != nil {
		_, err := base.Exec(
			ctx, query.UpdateConsigneePhoneQuery,
			consigneeID, req.Phone,
		)
		if err != nil {
			return err
		}
	}

	if req.Address != nil {
		a := req.Address
		_, err := base.Exec(
			ctx, query.UpdateConsigneeAddressQuery,
			consigneeID,
			a.Street,
			a.Number,
			a.Neighborhood,
			a.City,
			a.State,
			a.CEP,
			a.Complement,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func HasReport(
	ctx context.Context,
	base db.Queryable,
	consignee, report any,
) (bool, error) {
	return model.HasResource(
		ctx,
		base,
		query.ConsigneeHasReportQuery,
		consignee,
		report,
	)
}
