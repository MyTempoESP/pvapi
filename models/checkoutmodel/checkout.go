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
	sessionID string,
	req ConfirmRequest,
) (ConfirmResponse, error) {
	var resp ConfirmResponse

	tx, err := pool.Begin(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	checkoutID, uniqueID, done, err := checkIdempotency(
		ctx, tx, req.ExternalReference,
	)
	if err != nil {
		return resp, err
	}
	if done {
		return ConfirmResponse{UniqueID: uniqueID, Status: "paid", CheckoutID: checkoutID}, tx.Commit(ctx)
	}

	if err := validateStock(ctx, tx, req.Items); err != nil {
		return resp, err
	}

	value, err := decimal.NewFromString(req.ValuePaid)
	if err != nil {
		return resp, fmt.Errorf("invalid value_paid: %w", err)
	}

	checkoutID, uniqueID, err = writeCheckout(
		ctx, tx,
		sessionID,
		req.ExternalReference,
		value,
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

	resp.CheckoutID = checkoutID
	resp.UniqueID = uniqueID
	resp.Status = "paid"
	resp.QrCode = qr
	return resp, nil
}

func checkIdempotency(
	ctx context.Context,
	tx pgx.Tx,
	externalReference string,
) (checkoutID int, uniqueID string, done bool, err error) {
	err = tx.QueryRow(
		ctx,
		query.FindCheckoutByExternalRefQuery,
		externalReference,
	).Scan(&checkoutID, &uniqueID)

	if err == nil {
		return checkoutID, uniqueID, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", false, nil
	}
	return 0, "", false, fmt.Errorf("failed to check idempotency: %w", err)
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

func writeCheckout(
	ctx context.Context,
	tx pgx.Tx,
	sessionID, externalReference string,
	value decimal.Decimal,
	items []ConfirmItem,
) (checkoutID int, uniqueID string, err error) {
	totalItems := 0
	for _, item := range items {
		totalItems += item.Quantity
	}

	description := fmt.Sprintf(
		"Checkout for session %s — %d items",
		sessionID,
		totalItems,
	)

	err = tx.QueryRow(
		ctx,
		query.InsertCheckoutQuery,
		util.Args{
			"value":              value,
			"description":        description,
			"external_reference": externalReference,
		},
	).Scan(&checkoutID, &uniqueID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to insert checkout: %w", err)
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
		return 0, "", fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Insert per-shop financial breakdown
	_, err = tx.Exec(
		ctx,
		query.InsertCheckoutShopQuery,
		checkoutID,
		uniqueID,
	)
	if err != nil {
		return 0, "", fmt.Errorf("failed to insert checkout_shop: %w", err)
	}

	// Compute and store value_discrepancy
	_, err = tx.Exec(
		ctx,
		query.UpdateCheckoutDiscrepancyQuery,
		checkoutID,
	)
	if err != nil {
		return 0, "", fmt.Errorf("failed to update value_discrepancy: %w", err)
	}

	return checkoutID, uniqueID, nil
}
