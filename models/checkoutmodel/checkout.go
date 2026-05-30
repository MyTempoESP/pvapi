package checkoutmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/mail"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/qrcode"
)

type ConfirmItem struct {
	PackageID int `json:"package_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

type ConfirmRequest struct {
	ValuePaid         string        `json:"value_paid" binding:"required"`
	ExternalReference string        `json:"external_reference" binding:"required"`
	Items             []ConfirmItem `json:"items" binding:"required,min=1,dive"`
	ClientEmail       string        `json:"client_email"`
	ClientName        string        `json:"client_name"`
}

type PackageStock struct {
	Available int `db:"available"`
}

type ConfirmResponse struct {
	UniqueID string `json:"unique_id"`
	Status   string `json:"status"`
	QrCode   string `json:"qr_code"`
	CartID   int    `json:"-"`
}

var (
	ErrCartNotFound      = errors.New("cart not found or already checked out")
	ErrCartEmpty         = errors.New("cart has no items")
	ErrPackageNotFound   = errors.New("package not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

func Confirm(
	ctx context.Context,
	pool *pgxpool.Pool,
	sessionID string,
	req ConfirmRequest,
) (ConfirmResponse, error) {
	var resp ConfirmResponse

	tx, err := pool.Begin(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	uniqueID, done, err := checkIdempotency(
		ctx, tx, sessionID, req.ExternalReference,
	)
	if err != nil {
		return resp, err
	}
	if done {
		return ConfirmResponse{UniqueID: uniqueID, Status: "paid"}, tx.Commit(ctx)
	}

	cartID, itemCount, err := resolveCart(ctx, tx, sessionID)
	if err != nil {
		return resp, err
	}
	resp.CartID = cartID

	if err := validateStock(ctx, tx, req.Items); err != nil {
		return resp, err
	}

	value, err := decimal.NewFromString(req.ValuePaid)
	if err != nil {
		return resp, fmt.Errorf("invalid value_paid: %w", err)
	}

	uniqueID, err = writeCheckout(
		ctx, tx,
		cartID,
		sessionID,
		req.ExternalReference,
		value,
		itemCount,
		req.Items,
	)
	if err != nil {
		return resp, err
	}

	if err := tx.Commit(ctx); err != nil {
		return resp, fmt.Errorf("failed to commit transaction: %w", err)
	}

	qr, err := qrcode.Encode(uniqueID)
	if err != nil {
		return resp, fmt.Errorf("failed to generate qr code: %w", err)
	}

	resp.UniqueID = uniqueID
	resp.Status = "paid"
	resp.QrCode = qr
	return resp, nil
}

func checkIdempotency(
	ctx context.Context,
	tx pgx.Tx,
	sessionID, externalReference string,
) (uniqueID string, done bool, err error) {
	err = tx.QueryRow(
		ctx,
		query.FindCheckoutBySessionRefQuery,
		sessionID,
		externalReference,
	).Scan(&uniqueID)

	if err == nil {
		return uniqueID, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return "", false, fmt.Errorf("failed to check idempotency: %w", err)
}

func resolveCart(
	ctx context.Context,
	tx pgx.Tx,
	sessionID string,
) (cartID int, itemCount int, err error) {
	if err = tx.QueryRow(
		ctx,
		query.FindCartBySessionQuery,
		sessionID,
	).Scan(&cartID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrCartNotFound
			return
		}
		err = fmt.Errorf("failed to resolve cart: %w", err)
		return
	}

	if err = tx.QueryRow(
		ctx,
		query.CountCartItemsQuery,
		cartID,
	).Scan(&itemCount); err != nil {
		err = fmt.Errorf("failed to count cart items: %w", err)
		return
	}
	if itemCount == 0 {
		err = ErrCartEmpty
		return
	}

	return cartID, itemCount, nil
}

func validateStock(
	ctx context.Context,
	tx pgx.Tx,
	items []ConfirmItem,
) error {
	for _, item := range items {
		result, err := util.WithCollectExactlyOneRow[PackageStock](
			ctx, tx,
			query.CheckPackageAvailabilityQuery,
			item.PackageID,
		)
		if err != nil {
			return fmt.Errorf(
				"%w: package %d",
				ErrPackageNotFound,
				item.PackageID,
			)
		}

		if result.Available < item.Quantity {
			return fmt.Errorf(
				"%w: package %d (need %d, have %d)",
				ErrInsufficientStock,
				item.PackageID,
				item.Quantity,
				result.Available,
			)
		}
	}

	return nil
}

// pickupEmailRow is a single row from ResolvePickupEmailDataQuery.
type pickupEmailRow struct {
	PickupName  string `db:"pickup_name"`
	ProductName string `db:"product_name"`
	Quantity    int    `db:"quantity"`
}

// ResolvePickupEmailData looks up product names and pickup locations
// for the given checkout items and returns them grouped by pickup location.
func ResolvePickupEmailData(
	ctx context.Context,
	base db.Queryable,
	items []ConfirmItem,
) ([]mail.PickupGroup, error) {
	pkgIDs := make([]int, len(items))
	qtys := make([]int, len(items))
	for i, item := range items {
		pkgIDs[i] = item.PackageID
		qtys[i] = item.Quantity
	}

	rows, err := util.WithCollect[pickupEmailRow](
		ctx, base,
		query.ResolvePickupEmailDataQuery,
		pkgIDs,
		qtys,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve pickup email data: %w", err)
	}

	var groups []mail.PickupGroup
	var current *mail.PickupGroup

	for _, row := range rows {
		if current == nil || current.PickupName != row.PickupName {
			if current != nil {
				groups = append(groups, *current)
			}
			current = &mail.PickupGroup{
				PickupName: row.PickupName,
				Items:      make([]mail.PickupItem, 0),
			}
		}
		current.Items = append(current.Items, mail.PickupItem{
			ProductName: row.ProductName,
			Quantity:    row.Quantity,
		})
	}
	if current != nil {
		groups = append(groups, *current)
	}

	return groups, nil
}

func writeCheckout(
	ctx context.Context,
	tx pgx.Tx,
	cartID int,
	sessionID, externalReference string,
	value decimal.Decimal,
	itemCount int,
	items []ConfirmItem,
) (string, error) {
	description := fmt.Sprintf(
		"Checkout for session %s — %d items",
		sessionID,
		itemCount,
	)

	var uniqueID string
	err := tx.QueryRow(
		ctx,
		query.InsertCheckoutQuery,
		util.Args{
			"cart_id":            cartID,
			"value":              value,
			"description":        description,
			"external_reference": externalReference,
		},
	).Scan(&uniqueID)
	if err != nil {
		return "", fmt.Errorf("failed to insert checkout: %w", err)
	}

	var checkoutID int
	err = tx.QueryRow(
		ctx,
		query.FindCheckoutIDByUniqueIDQuery,
		uniqueID,
	).Scan(&checkoutID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve checkout id: %w", err)
	}

	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(
			query.InsertPackageReservedQuery,
			item.PackageID,
			checkoutID,
			item.Quantity,
		)
	}

	if err := util.Bake(ctx, tx, batch); err != nil {
		return "", fmt.Errorf("failed to reserve stock: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		query.FinishCartQuery,
		sessionID,
		externalReference,
	)
	if err != nil {
		return "", fmt.Errorf("failed to finish cart: %w", err)
	}

	return uniqueID, nil
}
