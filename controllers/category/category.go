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

		err = util.TransactionAware(
			c.Request.Context(), pool,
			func (ctx context.Context, tx pgx.Tx) error {
				return categorymodel.CreateTx(
					ctx, tx,
					request.Name,
					"",
					user.ActiveRoleID,
				)
			},
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "Categoria criada com sucesso")
	}
}

