package clientmodel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type Client struct {
	ID             int             `db:"user_id"`
	Name           string          `db:"user_name"`
	Email          string          `db:"user_email"`
	Phone          *string         `db:"user_phone"`
	TotalPurchases decimal.Decimal `db:"total_purchases"`
	LastPurchase   *time.Time      `db:"last_purchase"`
}

// SetCheckoutShopClientID sets the client_id on checkout_shop rows for the
// given checkout_id directly (no cart join).
func SetCheckoutShopClientID(
	ctx context.Context,
	tx pgx.Tx,
	clientID,
	checkoutID int,
) error {
	_, err := tx.Exec(
		ctx, query.UpdateCheckoutShopClientIDByCheckoutIDQuery,
		clientID,
		checkoutID,
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

// CreateAndAssociateByCheckout wraps Create plus writing client_id to
// checkouts, checkout_shop, and package_reserved tables using the
// given checkout_id. It opens its own transaction.
func CreateAndAssociateByCheckout(
	ctx context.Context,
	pool *pgxpool.Pool,
	clientName,
	clientEmail string,
	passwordHash []byte,
	checkoutID int,
) (isNew bool, err error) {
	err = util.TransactionAware(ctx, pool,
		func(ctx context.Context, tx pgx.Tx) error {
			clientID, isNewFlag, err := Create(ctx, tx, clientName, clientEmail, passwordHash)
			if err != nil {
				return err
			}
			isNew = isNewFlag
			if err := SetCheckoutClientID(ctx, tx, clientID, checkoutID); err != nil {
				return err
			}
			if err := SetCheckoutShopClientID(ctx, tx, clientID, checkoutID); err != nil {
				return err
			}
			if err := SetPackageReservedClientID(ctx, tx, clientID, checkoutID); err != nil {
				return err
			}
			return nil
		},
	)
	return
}

// SetCheckoutClientID sets client_id on the checkouts table.
func SetCheckoutClientID(
	ctx context.Context,
	tx pgx.Tx,
	clientID,
	checkoutID int,
) error {
	_, err := tx.Exec(
		ctx, query.UpdateCheckoutClientIDQuery,
		clientID,
		checkoutID,
	)
	return err
}

// SetPackageReservedClientID sets client_id on package_reserved rows for the
// given checkout_id.
func SetPackageReservedClientID(
	ctx context.Context,
	tx pgx.Tx,
	clientID,
	checkoutID int,
) error {
	_, err := tx.Exec(
		ctx, query.UpdatePackageReservedClientIDQuery,
		clientID,
		checkoutID,
	)
	return err
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
