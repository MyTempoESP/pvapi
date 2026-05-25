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

	packageMap, err := validateStock(ctx, tx, req.Items)
	if err != nil {
		return resp, err
	}

	value, err := decimal.NewFromString(req.ValuePaid)
	if err != nil {
		return resp, fmt.Errorf("invalid value_paid: %w", err)
	}

	// Persist checkout, reservations and finish cart
	uniqueID, err = writeCheckout(
		ctx, tx,
		cartID,
		sessionID,
		req.ExternalReference,
		value,
		itemCount,
		req.Items,
		packageMap,
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

	return ConfirmResponse{UniqueID: uniqueID, Status: "paid", QrCode: qr}, nil
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
) (map[int]int, error) {
	packageMap := make(map[int]int) // product_id -> package_id

	for _, item := range items {
		// TODO: duplicate packages for the same (product, consignment)
		// is a known data bug. As a workaround, pick the package
		// with the most available stock.
		candidates, err := util.WithCollect[PackageStock](
			ctx, tx,
			query.CheckPackageAvailabilityQuery,
			item.ProductID,
			item.ConsignmentID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to check stock: %w", err)
		}

		if len(candidates) == 0 {
			return nil, fmt.Errorf(
				"%w: product %d in consignment %d",
				ErrProductNotInConsignment,
				item.ProductID,
				item.ConsignmentID,
			)
		}

		best := candidates[0]
		for i := 1; i < len(candidates); i++ {
			if candidates[i].Available > best.Available {
				best = candidates[i]
			}
		}

		if best.Available < item.Quantity {
			return nil, fmt.Errorf(
				"%w: product %d (need %d, have %d)",
				ErrInsufficientStock,
				item.ProductID,
				item.Quantity,
				best.Available,
			)
		}

		packageMap[item.ProductID] = best.PackageID
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
