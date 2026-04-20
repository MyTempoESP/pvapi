package main

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/controllers/cart"
	"linhazero.com/pvapi/controllers/client"
	"linhazero.com/pvapi/controllers/consignee"
	"linhazero.com/pvapi/controllers/consignment"
	"linhazero.com/pvapi/controllers/pickups"
	"linhazero.com/pvapi/controllers/product"
	"linhazero.com/pvapi/controllers/report"
	"linhazero.com/pvapi/controllers/shop"
	"linhazero.com/pvapi/controllers/user"
	"linhazero.com/pvapi/controllers/payment"
	"linhazero.com/pvapi/middleware/auth"
)

func setupRoutes(
	t *gin.Engine,
	authMiddleware *jwt.GinJWTMiddleware,
	pool *pgxpool.Pool,
) {
	auth.RegisterAuthRoutes(t, authMiddleware, pool)

	withAuth := authMiddleware.MiddlewareFunc()

	// useful alias
	public := t
	t.POST("payments/session/create", payment.CreateTemporarySession)

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
		public.GET("/estabelecimentos", shop.ListHandler(pool))
		public.GET("/estabelecimentos/:shop/produtos",
			shop.PublicListProductsHandler(pool))

		public.POST("/estabelecimentos", shop.CreateHandler(pool))

		shops.GET(":shop", shop.ViewHandler(pool))
		shops.PUT(":shop", shop.UpdateHandler(pool))
	}

	products := t.Group("/produtos", withAuth)
	{
		public.GET("/produtos/:product",
			product.PublicViewHandler(pool))

		products.GET("/estabelecimento",
			product.ListByShopHandler(pool))
		products.PUT(":product", product.UpdateHandler(pool))
		products.POST("", product.CreateHandler(pool))

		products.GET(
			":product/destaque",
			product.UpdateHandler(pool),
		)

//		products.DELETE(
//			":product/img/:url",
//  			product.DeleteImageHandler(pool),
// 		)

	}

	consignees := t.Group("/monitores", withAuth)
	{
		consignees.POST("", consignee.CreateHandler(pool))
		consignees.GET("", consignee.ListHandler(pool))
	}

	consignments := t.Group("/remessas", withAuth)
	{
		consignments.POST("", consignment.CreateHandler(pool))
		consignments.GET("", consignment.ListHandler(pool))

		consignments.GET(":consignment", consignment.ViewHandler(pool))
		consignments.PUT(":consignment", consignment.UpdateHandler(pool))
		consignments.DELETE(":consignment", consignment.DeleteHandler(pool))

		consignments.GET(
			":consignment/produtos",
			consignment.ListProductsHandler(pool))

		consignments.POST(
			":consignment/produtos",
			consignment.AddProductHandler(pool))

		consignments.PUT(
			":consignment/produtos/:product",
			consignment.UpdateProductHandler(pool))
	}

	pickupLocations := t.Group("/pontos", withAuth)
	{
		pickupLocations.POST("", pickups.CreateHandler(pool))
		pickupLocations.GET("", pickups.ListHandler(pool))
	}

	reports := t.Group("/ocorrencias", withAuth)
	{
		reports.GET("", report.ListHandler(pool))
		reports.POST("", report.CreateHandler(pool))
		reports.PUT(":report", report.UpdateHandler(pool))
		reports.DELETE(":report", report.DeleteHandler(pool))
		reports.GET(":report", report.ViewHandler(pool))
	}

	clients := t.Group("/clientes", withAuth)
	{
		clients.POST("", client.CreateHandler(pool))
		clients.GET("", client.ListHandler(pool))
	}

	admin := t.Group("/admin", withAuth)
	{
		admin.GET(
			"/moderacao/pendentes",
			product.ListPendingModerationHandler(pool))

		admin.POST(
			"/moderacao/produtos/:product/aprovar",
			product.ModerationHandler(pool, "approved"),
		)

		admin.POST(
			"/moderacao/produtos/:product/rejeitar",
			product.ModerationHandler(pool, "rejected"),
		)
	}

	carts := t.Group("/carrinho")
	{
		carts.POST("/adicionar", cart.AddProductHandler(pool))
	}

	t.Static("/assets", "./public/uploads")
	t.Static("/public/uploads", "./public/uploads") // alias
}
