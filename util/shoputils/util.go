package shoputils

import (
	"context"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/shopmodel"
)

func FindFromContext(
	c *gin.Context,
	base db.Queryable,
) (*shopmodel.Shop, error) {
	user, err := authmodel.GetUserFromContext(c)

	ctx := c.Request.Context()

	if err != nil {
		return nil, fmt.Errorf("erro ao obter usuário logado: %w", err)
	}

	shop, err := shopmodel.FindByUser(ctx, base, user)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"loja não encontrada para usuário %q: %w", user, err)
		}

		return nil, fmt.Errorf(
			"erro ao buscar loja do usuário %q: %w", user, err)
	}

	return &shop, nil
}

// TODO FIXME: avoid repetition
func FindFromContextDetailed(
	c *gin.Context,
	base db.Queryable,
) (*shopmodel.DetailedShop, error) {
	user, err := authmodel.GetUserFromContext(c)

	ctx := c.Request.Context()

	if err != nil {
		return nil, fmt.Errorf("erro ao obter usuário logado: %w", err)
	}

	shop, err := shopmodel.FindByUserDetailed(ctx, base, user)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"loja não encontrada para usuário %q: %w", user, err)
		}

		return nil, fmt.Errorf(
			"erro ao buscar loja do usuário %q: %w", user, err)
	}

	return &shop, nil
}

func WithShopFromContextReturn[T any](
	c *gin.Context,
	base db.Queryable,
	provider func(context.Context, *shopmodel.Shop) (T, error),
) (T, error) {
	// empty value to return on failure
	var zero T

	shop, err := FindFromContext(c, base)

	if err != nil {
		return zero, err
	}

	return provider(c.Request.Context(), shop)
}
