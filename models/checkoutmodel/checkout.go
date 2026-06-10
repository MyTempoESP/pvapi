package checkoutmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/checkouthash"
	"linhazero.com/pvapi/util/qrcode"
)

type ConfirmItem struct {
	PackageID int `json:"package_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

type ConfirmRequest struct {
	ValuePaid   string        `json:"value_paid" binding:"required"`
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
)

func Confirm(
	ctx context.Context,
	pool *pgxpool.Pool,
	req ConfirmRequest,
	apiClientID int,
) (ConfirmResponse, error) {
	var resp ConfirmResponse

	hashItems := make([]checkouthash.Item, len(req.Items))
	for i, it := range req.Items {
		hashItems[i] = checkouthash.Item{PackageID: it.PackageID, Quantity: it.Quantity}
	}

	contentHash, err := checkouthash.Hash(req.ValuePaid, hashItems, req.ClientEmail)
	if err != nil {
		return resp, fmt.Errorf("failed to compute content hash: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	value, err := decimal.NewFromString(req.ValuePaid)
	if err != nil {
		return resp, fmt.Errorf("invalid value_paid: %w", err)
	}

	totalItems := 0
	for _, item := range req.Items {
		totalItems += item.Quantity
	}

	// Try INSERT with ON CONFLICT DO NOTHING — the idempotency mechanism.
	// If the same content_hash already exists, INSERT is a no-op.
	checkoutID, uniqueID, err := insertCheckoutContentHash(
		ctx, tx, value, contentHash, totalItems, apiClientID,
	)
	if err != nil {
		return resp, err
	}

	if checkoutID == 0 {
		// Duplicate content hash — fetch existing checkout
		checkoutID, uniqueID, err = findCheckoutByContentHash(ctx, tx, contentHash)
		if err != nil {
			return resp, err
		}

		qrEntries, err := findPackageReservedQRCodes(ctx, tx, checkoutID)
		if err != nil {
			return resp, err
		}

		resp.CheckoutID = checkoutID
		resp.UniqueID = uniqueID
		resp.Status = "paid"
		resp.QrCodes = qrEntries
		return resp, tx.Commit(ctx)
	}

	// New checkout — validate stock and write remaining data
	if err := validateStock(ctx, tx, req.Items); err != nil {
		return resp, err
	}

	qrEntries, err := writeCheckoutData(ctx, tx, checkoutID, uniqueID, req.Items)
	if err != nil {
		return resp, err
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



// insertCheckoutContentHash inserts a checkout with content-hash dedup.
// Returns (0, "", nil) when the content_hash already exists.
func insertCheckoutContentHash(
	ctx context.Context,
	tx pgx.Tx,
	value decimal.Decimal,
	contentHash string,
	totalItems int,
	apiClientID int,
) (checkoutID int, uniqueID string, err error) {
	description := buildDescription(totalItems)

	args := util.Merge(util.Args{
		"value":        value,
		"description":  description,
		"content_hash": contentHash,
	}, util.Args{
		"api_client_id": nilIfZero(apiClientID),
	})

	err = tx.QueryRow(
		ctx,
		query.InsertCheckoutQuery,
		args,
	).Scan(&checkoutID, &uniqueID)
	if errors.Is(err, pgx.ErrNoRows) {
		// ON CONFLICT DO NOTHING — no row returned, existing checkout
		return 0, "", nil
	}
	if err != nil {
		return 0, "", fmt.Errorf("failed to insert checkout: %w", err)
	}
	return checkoutID, uniqueID, nil
}

// findCheckoutByContentHash fetches an existing checkout by its content hash.
func findCheckoutByContentHash(
	ctx context.Context,
	tx pgx.Tx,
	contentHash string,
) (checkoutID int, uniqueID string, err error) {
	err = tx.QueryRow(
		ctx,
		query.FindCheckoutByContentHashQuery,
		contentHash,
	).Scan(&checkoutID, &uniqueID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to find existing checkout: %w", err)
	}
	return checkoutID, uniqueID, nil
}

// buildDescription returns the checkout description as a static label.
func buildDescription(totalItems int) string {
	return fmt.Sprintf("Checkout — %d items", totalItems)
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

	// Compute and store value_discrepancy
	_, err = tx.Exec(
		ctx,
		query.UpdateCheckoutDiscrepancyQuery,
		checkoutID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update value_discrepancy: %w", err)
	}

	return entries, nil
}

// findPackageReservedQRCodes fetches package_reserved rows for a checkout
// and generates QR codes for each one (dedup path).
func findPackageReservedQRCodes(
	ctx context.Context,
	tx pgx.Tx,
	checkoutID int,
) ([]QRCodeEntry, error) {
	rows, err := tx.Query(ctx, query.FindPackageReservedByCheckoutQuery, checkoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to query package_reserved for checkout %d: %w", checkoutID, err)
	}
	defer rows.Close()

	var entries []QRCodeEntry
	for rows.Next() {
		var uid string
		var pkgID int
		if err := rows.Scan(&uid, &pkgID); err != nil {
			return nil, fmt.Errorf("failed to scan package_reserved row: %w", err)
		}

		qr, err := qrcode.Encode(uid)
		if err != nil {
			return nil, fmt.Errorf("failed to generate qr code for package %d: %w", pkgID, err)
		}

		entries = append(entries, QRCodeEntry{
			PackageID: pkgID,
			QrCode:    qr,
		})
	}

	return entries, rows.Err()
}
