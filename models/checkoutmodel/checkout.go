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

	// Idempotency check
	var existingUUID string
	err = tx.QueryRow(
		ctx,
		query.FindCheckoutBySessionRefQuery,
		sessionID,
		req.ExternalReference,
	).Scan(&existingUUID)

	if err == nil {
		resp.UniqueID = existingUUID
		resp.Status = "paid"
		return resp, tx.Commit(ctx)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return resp, fmt.Errorf("failed to check idempotency: %w", err)
	}

	// Resolve cart_id from session_id
	var cartID int
	err = tx.QueryRow(
		ctx,
		`SELECT id FROM carts
		 WHERE session_id = $1
		   AND session_id NOT LIKE 'FINISHED-%'`,
		sessionID,
	).Scan(&cartID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, ErrCartNotFound
		}
		return resp, fmt.Errorf("failed to resolve cart: %w", err)
	}

	// Validate cart has items
	var itemCount int
	err = tx.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM cart_product WHERE cart_id = $1`,
		cartID,
	).Scan(&itemCount)
	if err != nil {
		return resp, fmt.Errorf("failed to count cart items: %w", err)
	}
	if itemCount == 0 {
		return resp, ErrCartEmpty
	}

	// Stock availability check
	packageMap := make(map[int]int) // product_id -> package_id
	for _, item := range req.Items {
		stock, err := util.WithCollectExactlyOneRow[PackageStock](
			ctx, tx,
			query.CheckPackageAvailabilityQuery,
			item.ProductID,
			item.ConsignmentID,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return resp, fmt.Errorf(
					"%w: product %d in consignment %d",
					ErrProductNotInConsignment,
					item.ProductID,
					item.ConsignmentID,
				)
			}
			return resp, fmt.Errorf("failed to check stock: %w", err)
		}

		if stock.Available < item.Quantity {
			return resp, fmt.Errorf(
				"%w: product %d (need %d, have %d)",
				ErrInsufficientStock,
				item.ProductID,
				item.Quantity,
				stock.Available,
			)
		}

		packageMap[item.ProductID] = stock.PackageID
	}

	// Parse value
	value, err := decimal.NewFromString(req.ValuePaid)
	if err != nil {
		return resp, fmt.Errorf("invalid value_paid: %w", err)
	}

	description := fmt.Sprintf(
		"Checkout for session %s — %d items",
		sessionID,
		itemCount,
	)

	// Insert checkout
	err = tx.QueryRow(
		ctx,
		query.InsertCheckoutQuery,
		util.Args{
			"cart_id":            cartID,
			"value":              value,
			"description":        description,
			"external_reference": req.ExternalReference,
		},
	).Scan(&resp.UniqueID)
	if err != nil {
		return resp, fmt.Errorf("failed to insert checkout: %w", err)
	}

	// Get checkout ID for package_reserved references
	var checkoutID int
	err = tx.QueryRow(
		ctx,
		`SELECT id FROM checkouts WHERE unique_id = $1`,
		resp.UniqueID,
	).Scan(&checkoutID)
	if err != nil {
		return resp, fmt.Errorf("failed to retrieve checkout id: %w", err)
	}

	// Reserve stock
	batch := &pgx.Batch{}
	for _, item := range req.Items {
		batch.Queue(
			query.InsertPackageReservedQuery,
			packageMap[item.ProductID],
			checkoutID,
			item.Quantity,
		)
	}

	err = util.Bake(ctx, tx, batch)
	if err != nil {
		return resp, fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Finish cart
	_, err = tx.Exec(
		ctx,
		query.FinishCartQuery,
		sessionID,
		req.ExternalReference,
	)
	if err != nil {
		return resp, fmt.Errorf("failed to finish cart: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to commit transaction: %w", err)
	}

	resp.Status = "paid"
	return resp, nil
}
