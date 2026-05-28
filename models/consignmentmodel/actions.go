package consignmentmodel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

type AddProductAction struct {
	Products      []AddProductUnit
	ConsignmentID int
}

type PackagingAttributes struct {
	Color string
}

type AddProductUnitAction struct {
	AttrVal PackagingAttributes
	AddProductUnit
}

func (u AddProductUnit) Prepare() AddProductUnitAction {
	return AddProductUnitAction{
		AddProductUnit: u,
		AttrVal: PackagingAttributes{
			Color: u.Color,
		},
	}
}

func (p AddProductRequest) Prepare(consignmentID int) AddProductAction {
	return AddProductAction{
		Products:      p.Products,
		ConsignmentID: consignmentID,
	}
}

type RemoveProductAction struct {
	ProductID     int
	ConsignmentID int
}

func (a RemoveProductAction) Exec(
	ctx context.Context,
	tx pgx.Tx,
) (err error) {
	if err = a.checkConsignmentNotArchived(ctx, tx); err != nil {
		return err
	}

	if err = a.checkNoActiveReservations(ctx, tx); err != nil {
		return err
	}

	currentQty, err := a.readPackageQuantity(ctx, tx)
	if err != nil {
		return err
	}

	if err = a.incrementProductStock(ctx, tx, currentQty); err != nil {
		return err
	}

	if err = a.removeFromPackages(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (a RemoveProductAction) checkConsignmentNotArchived(
	ctx context.Context,
	tx pgx.Tx,
) error {
	var archived int
	err := tx.QueryRow(
		ctx, query.CheckConsignmentArchivedQuery,
		a.ConsignmentID,
	).Scan(&archived)

	if err == nil {
		return fmt.Errorf(
			"não é possível remover produtos de uma remessa arquivada: %w",
			util.ErrBadRequest,
		)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf(
			"could not check consignment %d status: %w",
			a.ConsignmentID, err,
		)
	}

	return nil
}

func (a RemoveProductAction) checkNoActiveReservations(
	ctx context.Context,
	tx pgx.Tx,
) error {
	var reserved int
	err := tx.QueryRow(
		ctx, query.CheckPackageReservedForProductQuery,
		a.ProductID, a.ConsignmentID,
	).Scan(&reserved)

	if err == nil {
		return fmt.Errorf(
			"product %d in consignment %d has active reservations: %w",
			a.ProductID, a.ConsignmentID,
			util.ErrResourceLocked,
		)
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf(
			"could not check reservations for product %d in consignment %d: %w",
			a.ProductID, a.ConsignmentID, err,
		)
	}

	return nil
}

func (a RemoveProductAction) readPackageQuantity(
	ctx context.Context,
	tx pgx.Tx,
) (int, error) {
	var currentQty int
	err := tx.QueryRow(
		ctx, query.VerifyPackageQuantity,
		a.ProductID, a.ConsignmentID,
	).Scan(&currentQty)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf(
			"product %d not found in consignment %d: %w",
			a.ProductID, a.ConsignmentID,
			util.ErrNotFound,
		)
	}

	if err != nil {
		return 0, fmt.Errorf(
			"could not read package quantity for product %d in consignment %d: %w",
			a.ProductID, a.ConsignmentID, err,
		)
	}

	return currentQty, nil
}

func (a RemoveProductAction) incrementProductStock(
	ctx context.Context,
	tx pgx.Tx,
	qty int,
) error {
	_, err := tx.Exec(
		ctx, query.ProductAddToStock,
		a.ProductID, qty,
	)

	if err != nil {
		return fmt.Errorf(
			"could not increment stock for product %d: %w",
			a.ProductID, err,
		)
	}

	return nil
}

func (a RemoveProductAction) removeFromPackages(
	ctx context.Context,
	tx pgx.Tx,
) error {
	_, err := tx.Exec(
		ctx, query.DeleteProductFromConsignmentQuery,
		a.ProductID, a.ConsignmentID,
	)

	if err != nil {
		return fmt.Errorf(
			"could not remove product %d from consignment %d: %w",
			a.ProductID, a.ConsignmentID, err,
		)
	}

	return nil
}
