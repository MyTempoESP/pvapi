package main

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/controllers/consignee"
	"linhazero.com/pvapi/controllers/product"
	"linhazero.com/pvapi/controllers/shop"
	"linhazero.com/pvapi/controllers/user"
	"linhazero.com/pvapi/middleware/auth"
)

func setupRoutes(
	t *gin.Engine,
	authMiddleware *jwt.GinJWTMiddleware,
	pool *pgxpool.Pool,
) {
	auth.RegisterAuthRoutes(t, authMiddleware, pool)

	withAuth := authMiddleware.MiddlewareFunc()

	// no auth
	t.GET("/estabelecimentos/:shop/produtos", product.ListByParamHandler(pool))
	t.GET("/produtos/:id", product.ViewHandler(pool))

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
		shops.PUT(":shop", shop.UpdateHandler(pool))

	}

	products := t.Group("/produtos", withAuth)
	{
		products.GET("/estabelecimento", product.ListByShopHandler(pool))
		products.PUT(":id", product.UpdateHandler(pool))
		products.POST("", product.CreateHandler(pool))
	}

	consignees := t.Group("/monitores", withAuth)
	{
		consignees.POST("", consignee.CreateHandler(pool))
		consignees.GET("", consignee.ListHandler(pool))
	}

	consignments := t.Group("/remessas", withAuth)
	{
		consignments.POST("")
	}
}
