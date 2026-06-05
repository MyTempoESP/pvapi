package checkoutmodel

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/mail"
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
	ValuePaid         string        `json:"value_paid" binding:"required"`
	ExternalReference string        `json:"external_reference"`
	Items             []ConfirmItem `json:"items" binding:"required,min=1,dive"`
	ClientEmail       string        `json:"client_email"`
	ClientName        string        `json:"client_name"`
	Telephone         string        `json:"telephone"`
	CPF               string        `json:"cpf"`
}

type PackageStock struct {
	Available int `db:"available"`
}

type ConfirmResponse struct {
	UniqueID   string `json:"unique_id"`
	Status     string `json:"status"`
	QrCode     string `json:"qr_code"`
	CheckoutID int    `json:"-"`
}

var (
	ErrPackageNotFound   = errors.New("package not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

func Confirm(
	ctx context.Context,
	pool *pgxpool.Pool,
	req ConfirmRequest,
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
		ctx, tx, value, req.ExternalReference, contentHash, totalItems,
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

		qr, err := qrcode.Encode(uniqueID)
		if err != nil {
			return resp, fmt.Errorf("failed to generate qr code: %w", err)
		}

		resp.CheckoutID = checkoutID
		resp.UniqueID = uniqueID
		resp.Status = "paid"
		resp.QrCode = qr
		return resp, tx.Commit(ctx)
	}

	// New checkout — validate stock and write remaining data
	if err := validateStock(ctx, tx, req.Items); err != nil {
		return resp, err
	}

	if err := writeCheckoutData(ctx, tx, checkoutID, uniqueID, req.Items); err != nil {
		return resp, err
	}

	if err := tx.Commit(ctx); err != nil {
		return resp, fmt.Errorf("failed to commit transaction: %w", err)
	}

	qr, err := qrcode.Encode(uniqueID)
	if err != nil {
		return resp, fmt.Errorf("failed to generate qr code: %w", err)
	}

	resp.CheckoutID = checkoutID
	resp.UniqueID = uniqueID
	resp.Status = "paid"
	resp.QrCode = qr
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

// pickupEmailRow is a single row from ResolvePickupEmailDataQuery.
type pickupEmailRow struct {
	PickupName     string `db:"pickup_name"`
	ProductName    string `db:"product_name"`
	Quantity       int    `db:"quantity"`
	AddrStreet     string `db:"addrstreet"`
	AddrNumber     int    `db:"addrnumber"`
	AddrCity       string `db:"addrcity"`
	AddrState      string `db:"addrstate"`
	AddrCep        string `db:"addrcep"`
	AddrNeigh      string `db:"addrneigh"`
	AddrComplement string `db:"addrcomplement"`
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
				Address:    formatPickupAddress(row),
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

// formatPickupAddress formats a pickupEmailRow's address fields into a
// single-line address string. Returns nil when no address data is present.
func formatPickupAddress(row pickupEmailRow) *mail.PickupAddress {
	if row.AddrStreet == "" {
		return nil
	}

	var b strings.Builder
	b.WriteString(row.AddrStreet)
	b.WriteString(", ")
	fmt.Fprint(&b, row.AddrNumber)

	if row.AddrComplement != "" && row.AddrComplement != "nenhum" {
		b.WriteString(", ")
		b.WriteString(row.AddrComplement)
	}

	b.WriteString(" — ")
	if row.AddrNeigh != "" {
		b.WriteString(row.AddrNeigh)
		b.WriteString(", ")
	}
	b.WriteString(row.AddrCity)
	b.WriteString("/")
	b.WriteString(row.AddrState)

	b.WriteString(" — CEP ")
	b.WriteString(row.AddrCep)

	return &mail.PickupAddress{Formatted: b.String()}
}

// insertCheckoutContentHash inserts a checkout with content-hash dedup.
// Returns (0, "", nil) when the content_hash already exists.
func insertCheckoutContentHash(
	ctx context.Context,
	tx pgx.Tx,
	value decimal.Decimal,
	externalReference string,
	contentHash string,
	totalItems int,
) (checkoutID int, uniqueID string, err error) {
	description := buildDescription(externalReference, totalItems)

	err = tx.QueryRow(
		ctx,
		query.InsertCheckoutQuery,
		util.Args{
			"value":              value,
			"description":        description,
			"external_reference": externalReference,
			"content_hash":       contentHash,
		},
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

// buildDescription returns the checkout description.
// Falls back to a generic label when external_reference is empty.
func buildDescription(externalReference string, totalItems int) string {
	reference := externalReference
	if reference == "" {
		reference = "Checkout"
	}
	return fmt.Sprintf("%s — %d items", reference, totalItems)
}

// writeCheckoutData writes the remaining checkout data (package reservations,
// financial breakdown, discrepancy) after a new checkout row is inserted.
func writeCheckoutData(
	ctx context.Context,
	tx pgx.Tx,
	checkoutID int,
	uniqueID string,
	items []ConfirmItem,
) error {
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
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Insert per-shop financial breakdown
	_, err := tx.Exec(
		ctx,
		query.InsertCheckoutShopQuery,
		checkoutID,
		uniqueID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert checkout_shop: %w", err)
	}

	// Compute and store value_discrepancy
	_, err = tx.Exec(
		ctx,
		query.UpdateCheckoutDiscrepancyQuery,
		checkoutID,
	)
	if err != nil {
		return fmt.Errorf("failed to update value_discrepancy: %w", err)
	}

	return nil
}
