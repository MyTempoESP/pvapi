package clientmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type Client struct {
	ID    int     `db:"user_id"`
	Name  string  `db:"user_name"`
	Email string  `db:"user_email"`
	Phone *string `db:"user_phone"`
}

// AssociateToCart sets the client_id on a cart and marks the session as finished.
// This is the original consignment-flow association.
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

// SetCartClientID sets only the client_id on a cart without touching session_id.
// Used for post-checkout client association.
func SetCartClientID(
	ctx context.Context,
	tx pgx.Tx,
	clientID,
	cartID int,
) error {
	_, err := tx.Exec(
		ctx, query.UpdateCartClientIDQuery,
		clientID,
		cartID,
	)
	return err
}

// Create creates a user and client record in a single transaction.
// - If the email already belongs to a client, returns the existing client ID with isNew=false.
// - If the email belongs to a non-client user, returns ErrDuplicatedResource.
// - The password hash is pre-computed (bcrypt is heavy — hash outside the tx).
func Create(
	ctx context.Context,
	tx pgx.Tx,
	clientName,
	clientEmail string,
	passwordHash []byte,
) (clientID int, isNew bool, err error) {
	userID, err := usermodel.FindByEmail(ctx, tx, clientEmail)

	if errors.Is(err, pgx.ErrNoRows) {
		userID, err = usermodel.Create(
			ctx, tx,
			clientName,
			clientEmail,
			passwordHash,
		)
		if err != nil {
			return 0, false, fmt.Errorf("error creating user: %w", err)
		}

		err = tx.QueryRow(
			ctx, query.CreateClientQuery,
			userID,
			nil, // telephone is not captured during checkout
		).Scan(&clientID)
		if err != nil {
			return 0, false, fmt.Errorf("error creating client: %w", err)
		}

		return clientID, true, nil
	}

	if err != nil {
		return 0, false, err
	}

	// email belongs to an existing user — check if they are a client
	err = tx.QueryRow(ctx, query.FindClientByUserQuery, userID).Scan(&clientID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("user is not a client: %w", util.ErrDuplicatedResource)
	}

	if err != nil {
		return 0, false, err
	}

	return clientID, false, nil
}

// CreateAndAssociate wraps Create plus setting the cart's client_id.
// It opens its own transaction, independent of any caller transaction.
func CreateAndAssociate(
	ctx context.Context,
	pool *pgxpool.Pool,
	clientName,
	clientEmail string,
	passwordHash []byte,
	cartID int,
) (isNew bool, err error) {
	err = util.TransactionAware(ctx, pool,
		func(ctx context.Context, tx pgx.Tx) error {
			clientID, isNewFlag, err := Create(ctx, tx, clientName, clientEmail, passwordHash)
			if err != nil {
				return err
			}
			isNew = isNewFlag
			return SetCartClientID(ctx, tx, clientID, cartID)
		},
	)
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

// TODO: archive instead of delete
func Delete(
	ctx context.Context,
	base db.Queryable,
	id any,
) error {
	_, err := base.Exec(ctx, query.DeleteClientQuery, id)
	return err
}
