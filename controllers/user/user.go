package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/models/authmodel"
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
func PermsClient(c *gin.Context) {
	c.JSON(200, gin.H{"status": "Permitido"})
}

func UserPermissionLevelHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)
		c.JSON(200, gin.H{"level": user.Role})
	}
}

func UserDataHandler(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		var userName, userMail string

		err := pool.QueryRow(
			c.Request.Context(),
			query.UserDataQuery, user.UserID,
		).Scan(&userName, &userMail)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		c.JSON(200, gin.H{
			"name":    userName,
			"email":   userMail,
			"message": "succesfully fetched user",
		})
	}
}
