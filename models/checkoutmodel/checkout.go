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
	"linhazero.com/pvapi/util/qrcode"
)

type ConfirmItem struct {
	ProductID     int `json:"product_id" binding:"required"`
	ConsignmentID int `json:"consignment_id" binding:"required"`
	Quantity      int `json:"quantity" binding:"required,min=1"`
}

type ConfirmRequest struct {
	ValuePaid         string        `json:"value_paid" binding:"required"`
	ExternalReference string        `json:"external_reference" binding:"required"`
	Items             []ConfirmItem `json:"items" binding:"required,min=1,dive"`
}

type PackageStock struct {
	PackageID       int `db:"package_id"`
	TotalAllocated  int `db:"total_allocated"`
	AlreadyReserved int `db:"already_reserved"`
	Available       int `db:"available"`
}

type ConfirmResponse struct {
	UniqueID string `json:"unique_id"`
	Status   string `json:"status"`
	QrCode   string `json:"qr_code"`
}

var (
	ErrCartNotFound      = errors.New("cart not found or already checked out")
	ErrCartEmpty         = errors.New("cart has no items")
	ErrProductNotInConsignment = errors.New("product not available in consignment")
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

	// 1. Idempotency
	uniqueID, done, err := checkIdempotency(ctx, tx, sessionID, req.ExternalReference)
	if err != nil {
		return resp, err
	}
	if done {
		return ConfirmResponse{UniqueID: uniqueID, Status: "paid"}, tx.Commit(ctx)
	}

	// 2. Cart resolution
	cartID, itemCount, err := resolveCart(ctx, tx, sessionID)
	if err != nil {
		return resp, err
	}

	// 3. Stock validation
	packageMap, err := validateStock(ctx, tx, req.Items)
	if err != nil {
		return resp, err
	}

	// 4. Parse payment value
	value, err := decimal.NewFromString(req.ValuePaid)
	if err != nil {
		return resp, fmt.Errorf("invalid value_paid: %w", err)
	}

	// 5. Persist checkout, reservations and finish cart
	uniqueID, err = writeCheckout(ctx, tx, cartID, sessionID, req.ExternalReference, value, itemCount, req.Items, packageMap)
	if err != nil {
		return resp, err
	}

	if err := tx.Commit(ctx); err != nil {
		return resp, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 6. QR code
	qr, err := qrcode.Encode(uniqueID)
	if err != nil {
		return resp, fmt.Errorf("failed to generate qr code: %w", err)
	}

	return ConfirmResponse{UniqueID: uniqueID, Status: "paid", QrCode: qr}, nil
}

func checkIdempotency(ctx context.Context, tx pgx.Tx, sessionID, externalReference string) (uniqueID string, done bool, err error) {
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

func resolveCart(ctx context.Context, tx pgx.Tx, sessionID string) (cartID int, itemCount int, err error) {
	err = tx.QueryRow(
		ctx,
		`SELECT id FROM carts
		 WHERE session_id = $1
		   AND session_id NOT LIKE 'FINISHED-%'`,
		sessionID,
	).Scan(&cartID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, ErrCartNotFound
		}
		return 0, 0, fmt.Errorf("failed to resolve cart: %w", err)
	}

	err = tx.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM cart_product WHERE cart_id = $1`,
		cartID,
	).Scan(&itemCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count cart items: %w", err)
	}
	if itemCount == 0 {
		return 0, 0, ErrCartEmpty
	}

	return cartID, itemCount, nil
}

func validateStock(ctx context.Context, tx pgx.Tx, items []ConfirmItem) (map[int]int, error) {
	packageMap := make(map[int]int) // product_id -> package_id

	for _, item := range items {
		stock, err := util.WithCollectExactlyOneRow[PackageStock](
			ctx, tx,
			query.CheckPackageAvailabilityQuery,
			item.ProductID,
			item.ConsignmentID,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf(
					"%w: product %d in consignment %d",
					ErrProductNotInConsignment,
					item.ProductID,
					item.ConsignmentID,
				)
			}
			return nil, fmt.Errorf("failed to check stock: %w", err)
		}

		if stock.Available < item.Quantity {
			return nil, fmt.Errorf(
				"%w: product %d (need %d, have %d)",
				ErrInsufficientStock,
				item.ProductID,
				item.Quantity,
				stock.Available,
			)
		}

		packageMap[item.ProductID] = stock.PackageID
	}

	return packageMap, nil
}

func writeCheckout(
	ctx context.Context,
	tx pgx.Tx,
	cartID int,
	sessionID, externalReference string,
	value decimal.Decimal,
	itemCount int,
	items []ConfirmItem,
	packageMap map[int]int,
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
		`SELECT id FROM checkouts WHERE unique_id = $1`,
		uniqueID,
	).Scan(&checkoutID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve checkout id: %w", err)
	}

	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(
			query.InsertPackageReservedQuery,
			packageMap[item.ProductID],
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
