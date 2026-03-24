package shoputils

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/shopmodel"
)

func FindFromContext(
	c *gin.Context,
	pool *pgxpool.Pool,
) (*shopmodel.Shop, error) {
	user, err := authmodel.GetUserFromContext(c)

	ctx := c.Request.Context()

	if err != nil {
		return nil, fmt.Errorf("erro ao obter usuário logado: %w", err)
	}

	shop, err := shopmodel.FindByUser(ctx, pool, user)

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
