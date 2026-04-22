package category

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"linhazero.com/pvapi/models/categorymodel"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/util"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		categories, err := categorymodel.ListByShop(
			c.Request.Context(), pool,
			user.ActiveRoleID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.Wrap(
			c, "Categorias listadas com sucesso",
			util.ResourceCollection(categories),
		)
	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request struct{
			Name string `json:"name"`
		}

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		user := authmodel.GetUserFromContext(c)

		var categoryID int

		err = util.TransactionAware(
			c.Request.Context(), pool,
			func (ctx context.Context, tx pgx.Tx) error {
				id, err := categorymodel.CreateTx(
					ctx, tx,
					request.Name,
					"",
					user.ActiveRoleID,
				)
				categoryID = id
				return err
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c, "Categoria criada com sucesso",
			"id", categoryID,
		)
	}
}

