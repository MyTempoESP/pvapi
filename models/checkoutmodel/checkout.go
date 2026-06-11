package checkoutmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/clientmodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/checkoutclient"
	"linhazero.com/pvapi/util/qrcode"
)

// RateLimitWindowSeconds is the cooldown period between checkouts
// for the same client, in seconds. Exported so tests can override.
var RateLimitWindowSeconds = 60

type ConfirmItem struct {
	PackageID int `json:"package_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

type ConfirmRequest struct {
	Items       []ConfirmItem `json:"items" binding:"required,min=1,dive"`
	ClientEmail string        `json:"client_email"`
	ClientName  string        `json:"client_name"`
	Telephone   string        `json:"telephone"`
	CPF         string        `json:"cpf"`
}

type PackageStock struct {
	Available int `db:"available"`
}

type QRCodeEntry struct {
	PackageID int    `json:"package_id"`
	QrCode    string `json:"qr_code"`
}

type ConfirmResponse struct {
	UniqueID   string        `json:"unique_id"`
	Status     string        `json:"status"`
	QrCodes    []QRCodeEntry `json:"qr_codes"`
	CheckoutID int           `json:"-"`
}

var (
	ErrPackageNotFound   = errors.New("package not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrRateLimited       = errors.New("rate limited")
	ErrNoIdentity        = errors.New("no identity provided")
)

// RateLimitInfo holds the rate-limit error details for the API response.
type RateLimitInfo struct {
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after"`
}

func (e *RateLimitInfo) Error() string {
	return e.Message
}

func Confirm(
	ctx context.Context,
	pool *pgxpool.Pool,
	req ConfirmRequest,
	apiClientID int,
) (ConfirmResponse, error) {
	var resp ConfirmResponse

	ci := checkoutclient.Resolve(
		req.ClientEmail,
		req.Telephone,
		req.CPF,
		req.ClientName,
	)
	if ci.EffectiveEmail == "" {
		return resp, fmt.Errorf("%w: no identity resolved", ErrNoIdentity)
	}

	passwordHash, err := util.HashPassword(ci.Password)
	if err != nil {
		return resp, fmt.Errorf("failed to hash password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Create or get client (synchronous, inside transaction)
	clientID, _, err := clientmodel.Create(
		ctx, tx,
		ci.ClientName, ci.EffectiveEmail,
		passwordHash,
		ci.NormalizedPhone,
		ci.ValidatedCPF,
	)
	if err != nil {
		return resp, fmt.Errorf("failed to create/resolve client: %w", err)
	}

	// For existing clients, update only the fields that were actually provided
	if ci.NormalizedPhone != "" || ci.ValidatedCPF != "" {
		if _, err := tx.Exec(ctx, query.UpdateClientFieldsQuery, clientID, ci.NormalizedPhone, ci.ValidatedCPF); err != nil {
			return resp, fmt.Errorf("error updating existing client fields: %w", err)
		}
	}

	// Serialize checkouts for this client with row-level lock
	_, err = tx.Exec(ctx, query.LockClientRowQuery, clientID)
	if err != nil {
		return resp, fmt.Errorf("failed to lock client row: %w", err)
	}

	// Rate-limit check: if last checkout within 60 seconds, return 429
	remaining, isBlocked, err := checkRateLimit(ctx, tx, clientID)
	if err != nil {
		return resp, err
	}
	if isBlocked {
		return resp, &RateLimitInfo{
			Message:    fmt.Sprintf("Aguarde %d segundos antes de fazer outra compra", remaining),
			RetryAfter: remaining,
		}
	}

	totalItems := 0
	for _, item := range req.Items {
		totalItems += item.Quantity
	}

	description := fmt.Sprintf("Checkout — %d items", totalItems)

	checkoutID, uniqueID, err := insertCheckout(
		ctx, tx, description, apiClientID, clientID,
	)
	if err != nil {
		return resp, err
	}

	if err := validateStock(ctx, tx, req.Items); err != nil {
		return resp, err
	}

	qrEntries, err := writeCheckoutData(ctx, tx, checkoutID, uniqueID, req.Items)
	if err != nil {
		return resp, err
	}

	// Set client_id on related tables
	if err := clientmodel.SetCheckoutClientID(ctx, tx, clientID, checkoutID); err != nil {
		return resp, fmt.Errorf("failed to set checkout client_id: %w", err)
	}
	if err := clientmodel.SetCheckoutShopClientID(ctx, tx, clientID, checkoutID); err != nil {
		return resp, fmt.Errorf("failed to set checkout_shop client_id: %w", err)
	}
	if err := clientmodel.SetPackageReservedClientID(ctx, tx, clientID, checkoutID); err != nil {
		return resp, fmt.Errorf("failed to set package_reserved client_id: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return resp, fmt.Errorf("failed to commit transaction: %w", err)
	}

	resp.CheckoutID = checkoutID
	resp.UniqueID = uniqueID
	resp.Status = "paid"
	resp.QrCodes = qrEntries
	return resp, nil
}

// checkRateLimit queries the most recent checkout for this client and
// returns the remaining seconds and whether the request is blocked.
func checkRateLimit(
	ctx context.Context,
	tx pgx.Tx,
	clientID int,
) (remaining int, isBlocked bool, err error) {
	var secondsAgo *int64
	err = tx.QueryRow(ctx, query.CheckRateLimitQuery, clientID).Scan(&secondsAgo)
	if errors.Is(err, pgx.ErrNoRows) {
		// No previous checkout — not blocked
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if secondsAgo == nil || *secondsAgo >= int64(RateLimitWindowSeconds) {
		return 0, false, nil
	}

	remaining = RateLimitWindowSeconds - int(*secondsAgo)
	return remaining, true, nil
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

// insertCheckout inserts a new checkout row and returns its ID and unique_id.
func insertCheckout(
	ctx context.Context,
	tx pgx.Tx,
	description string,
	apiClientID int,
	clientID int,
) (checkoutID int, uniqueID string, err error) {
	args := util.Merge(
		util.Args{
			"description": description,
		},
		util.Args{
			"api_client_id": nilIfZero(apiClientID),
			"client_id":     clientID,
		},
	)

	err = tx.QueryRow(ctx, query.InsertCheckoutQuery, args).Scan(&checkoutID, &uniqueID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to insert checkout: %w", err)
	}
	return checkoutID, uniqueID, nil
}

// nilIfZero converts a zero int to nil for nullable DB columns.
func nilIfZero(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

// writeCheckoutData writes package reservations, generates QR codes per
// package, and records financial breakdown.
func writeCheckoutData(
	ctx context.Context,
	tx pgx.Tx,
	checkoutID int,
	uniqueID string,
	items []ConfirmItem,
) ([]QRCodeEntry, error) {
	var entries []QRCodeEntry

	for _, item := range items {
		var pkgUniqueID string
		var pkgID int

		err := tx.QueryRow(
			ctx,
			query.InsertPackageReservedQuery,
			item.PackageID,
			checkoutID,
			item.Quantity,
		).Scan(&pkgUniqueID, &pkgID)
		if err != nil {
			return nil, fmt.Errorf("failed to reserve package %d: %w", item.PackageID, err)
		}

		qr, err := qrcode.Encode(pkgUniqueID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate qr code for package %d: %w", pkgID, err)
		}

		entries = append(entries, QRCodeEntry{
			PackageID: pkgID,
			QrCode:    qr,
		})
	}

	// Insert per-shop financial breakdown
	_, err := tx.Exec(
		ctx,
		query.InsertCheckoutShopQuery,
		checkoutID,
		uniqueID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert checkout_shop: %w", err)
	}

	return entries, nil
}
