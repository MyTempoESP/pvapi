package user

import (
	"github.com/gin-gonic/gin"
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

func UserPermissionsHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		perms, err := usermodel.GetUserPermissions(c.Request.Context(), pool, user.UserID)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.Wrap(c, "permissoes obtidas", perms)
	}
}

func PermissionsByEmailHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			c.JSON(400, gin.H{"error": "email query parameter required"})
			return
		}

		userID, err := usermodel.FindByEmail(c.Request.Context(), pool, email)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		perms, err := usermodel.GetUserPermissions(c.Request.Context(), pool, userID)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.Wrap(c, "permissoes obtidas", perms)
	}
}
