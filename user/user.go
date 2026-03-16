package user

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

func PermsAdmin(c *gin.Context) {
	c.JSON(200, gin.H{"status": "Permitido"})
}
func PermsShop(c *gin.Context) {
	c.JSON(200, gin.H{"status": "Permitido"})
}
func PermsConsignee(c *gin.Context) {
	c.JSON(200, gin.H{"status": "Permitido"})
}

func UserPermissionLevelHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, err := authmodel.GetUserFromContext(c)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		level, err := usermodel.GetUserPermissionLevel(
			c.Request.Context(),
			pool, user,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.Handle404(c)
				return
			}

			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{"level": level})
	}
}

func UserDataHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, err := authmodel.GetUserFromContext(c)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		var userName, userMail string

		err = pool.QueryRow(
			c.Request.Context(),
			query.UserDataQuery, user,
		).Scan(&userName, &userMail)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.Handle404(c)
				return
			}

			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"name":    userName,
			"email":   userMail,
			"message": "succesfully fetched user",
		})
	}
}
