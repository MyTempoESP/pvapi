package user

import (
	"errors"
	"fmt"
	"strconv"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/auth"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

func UserDataHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var err error

		claims := jwt.ExtractClaims(c)
		authUser := auth.GetUserFromClaims(claims)

		userIDString := authUser.UserID

		user, err := strconv.Atoi(userIDString)

		// wrapper for custom error instead of postgres screaming
		if err != nil {
			util.AbortWithInternalError(
				c,
				fmt.Errorf("failed to convert user id %q", userIDString),
			)
			return
		}

		var userName, userMail string

		err = pool.QueryRow(
			c.Request.Context(),
			query.UserDataQuery, user,
		).Scan(&userName, &userMail)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				util.AbortWithInternalError(
					c,
					fmt.Errorf("user %q not found", userIDString),
				)
				return
			}

			util.AbortWithInternalError(c, err)
			return
		}

		c.JSON(200, gin.H{
			"name":    userName,
			"email":   userMail,
			"message": "succesfully fetched user",
		})
	}
}
