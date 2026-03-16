package main

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/auth"
	"linhazero.com/pvapi/shop"
	"linhazero.com/pvapi/user"
)

func setupRoutes(
	t *gin.Engine,
	authMiddleware *jwt.GinJWTMiddleware,
	pool *pgxpool.Pool,
) {
	auth.RegisterAuthRoutes(t, authMiddleware, pool)

	withAuth := authMiddleware.MiddlewareFunc()

	loggedIn := t.Group("", withAuth)
	{
		loggedIn.GET("/logout", authMiddleware.LogoutHandler)
		loggedIn.GET("/user", user.UserDataHandler(pool))
		loggedIn.GET("/level", user.UserPermissionLevelHandler(pool))
	}

	perms := t.Group("/perms", withAuth)
	{
		perms.GET("/admin", user.PermsAdmin)
		perms.GET("/shop", user.PermsShop)
		perms.GET("/monitor", user.PermsConsignee)
	}

	shops := t.Group("/estabelecimentos", withAuth)
	{
		shops.GET("", shop.ListHandler(pool))
		shops.POST("", shop.CreateHandler(pool))

		shops.GET(":shop", shop.ViewHandler(pool))
	}
}
