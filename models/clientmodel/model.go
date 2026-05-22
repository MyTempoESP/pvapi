package clientmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/userutils"
)


type Client struct {
	ID    int    `db:"user_id"`
	Name  string `db:"user_name"`
	Email string `db:"user_email"`
	Phone string `db:"user_phone"`
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
	userID, err := usermodel.FindByEmail(ctx, tx, clientEmail)

	if errors.Is(err, pgx.ErrNoRows) {
		user, err := userutils.NewUser(
			clientName,
			clientEmail,
			clientPhone,
		)
		if err != nil {
			return err
		}

		userID, err = usermodel.Create(
			ctx, tx,
			clientName,
			clientEmail,
			user.PasswordHash,
		)
		if err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}

		var clientID int
		err = tx.QueryRow(
			ctx, query.CreateClientQuery,
			userID,
			clientPhone,
		).Scan(&clientID)
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		return AssociateToCart(ctx, tx, clientID, cartID)
	}

	if err != nil {
		return err
	}

	var clientID int
	err = tx.QueryRow(ctx, query.FindClientByUserQuery, userID).Scan(&clientID)

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("user is not a client: %w", util.ErrDuplicatedResource)
	}

	if err != nil {
		return err
	}

	return AssociateToCart(ctx, tx, clientID, cartID)
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

// TODO: archive instead of delete
func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteClientQuery, id)
	return err
}
