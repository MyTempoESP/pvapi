package clientmodel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/userutils"

	sq "github.com/Masterminds/squirrel"
)

type Client struct {
	ID    int
	Field any
}

func (c Client) Save(
	ctx context.Context,
	base db.Queryable,
) (err error) {
	q := util.SQL

	fields := make(map[string]any)

	if c.Field != nil {
		fields["field"] = c.Field
	}

	if len(fields) == 0 {
		return
	}

	query, params, err := q.Update("clients").
		SetMap(fields).
		Where(sq.Eq{"id": c.ID}).
		ToSql()

	if err != nil {
		return fmt.Errorf("could not build SQL query: %w", err)
	}

	_, err = base.Exec(ctx, query, params...)

	if err != nil {
		return fmt.Errorf(
			"could not update client, query: %q, %w", query, err)
	}

	return
}

func AssociateToCart(
	ctx context.Context,
	tx pgx.Tx,
	clientID,
	cartID int,
) error {
	_, err := tx.Exec(
		ctx, query.ClientAssociateToCartQuery,
		clientID,
		cartID,
	)
	return err
}

func CreateForCart(
	ctx context.Context,
	tx pgx.Tx,
	clientName,
	clientEmail,
	clientPhone string, // password
	cartID int,
) (err error) {
	// XXX: not a good idea to generate
	// a password inside a transaction
	user, err := userutils.NewUser(
		clientName,
		clientEmail,
		clientPhone,
	)

	if err != nil {
		return
	}

	userID, err := user.Save(ctx, tx)

	if err != nil {
		return
	}

	var clientID int

	err = tx.QueryRow(
		ctx, query.CreateClientQuery,
		userID,
	).Scan(&clientID)

	if err != nil {
		return
	}

	err = AssociateToCart(ctx, tx, clientID, cartID)

	return
}

func List(
	ctx context.Context,
	base db.Queryable,
	shopID int,
) ([]Client, error) {
	return util.WithCollect[Client](
		ctx,
		base,
		query.ListShopClientsQuery,
		shopID,
	)
}

func Find(
	ctx context.Context,
	base db.Queryable,
	id any,
) (Client, error) {
	return util.WithCollectExactlyOneRow[Client](
		ctx,
		base,
		query.FindClientQuery,
		id,
	)
}

func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteClientQuery, id)
	return err
}
