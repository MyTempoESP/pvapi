package main

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/controllers/bankaccount"
	"linhazero.com/pvapi/controllers/cart"
	"linhazero.com/pvapi/controllers/category"
	"linhazero.com/pvapi/controllers/checkout"
	"linhazero.com/pvapi/controllers/client"
	"linhazero.com/pvapi/controllers/consignee"
	"linhazero.com/pvapi/controllers/consignment"
	"linhazero.com/pvapi/controllers/payment"
	"linhazero.com/pvapi/controllers/pickups"
	"linhazero.com/pvapi/controllers/product"
	"linhazero.com/pvapi/controllers/report"
	"linhazero.com/pvapi/controllers/saque"
	"linhazero.com/pvapi/controllers/shop"
	"linhazero.com/pvapi/controllers/user"
	"linhazero.com/pvapi/middleware/auth"
	"linhazero.com/pvapi/util"
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

	util.Sybau(public)

	public.GET("/permissions", user.PermissionsByEmailHandler(pool))

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
		shops.GET("/all", shop.ListRestrictedHandler(pool))
		shops.GET("/checkouts", shop.CheckoutsReportHandler(pool))
		shops.POST("/saques", saque.CreateHandler(pool))
		shops.GET("/saques", saque.ListHandler(pool))
		shops.GET("/saques/:saque", saque.ViewHandler(pool))
		shops.PUT("/saques/:saque", saque.UpdateStatusHandler(pool))
		shops.DELETE("/saques/:saque", saque.DeleteHandler(pool))
		shops.POST("/contasbancarias", bankaccount.CreateHandler(pool))
		shops.GET("/contasbancarias", bankaccount.ListHandler(pool))
		shops.PUT("/contasbancarias/:id", bankaccount.UpdateHandler(pool))
		shops.DELETE("/contasbancarias/:id", bankaccount.DeleteHandler(pool))
		shops.GET(":shop", shop.ViewHandler(pool))
		shops.PUT(":shop", shop.UpdateHandler(pool))
		shops.DELETE(":shop", shop.DeactivateHandler(pool))

		public.GET("/estabelecimentos", shop.ListHandler(pool))
		public.GET("/estabelecimentos/:shop/produtos",
			shop.PublicListProductsHandler(pool))
		public.GET("/estabelecimentos/:shop/remessas",
			shop.PublicListConsignmentsHandler(pool))

		public.POST("/estabelecimentos", shop.CreateHandler(pool))
	}

	categories := t.Group("/categorias", withAuth)
	{
		categories.POST("", category.CreateHandler(pool))
		categories.GET("", category.ListHandler(pool))
		// TOOD: make a public version
	}

	products := t.Group("/produtos", withAuth)
	{
		public.GET("/produtos/:product",
			product.PublicViewHandler(pool))

		public.GET("/produtos",
			product.PublicListingHandler(pool))

		products.GET("/estabelecimento",
			product.ListByShopHandler(pool))
		products.PUT(":product", product.UpdateHandler(pool))
		products.POST("", product.CreateHandler(pool))

		products.GET(
			":product/destaque",
			product.UpdateHandler(pool),
		)

		products.DELETE(":product", product.DeleteHandler(pool))

		products.DELETE(
			":product/img/:url",
			product.DeleteImageHandler(pool),
		)
	}

	consignees := t.Group("/monitores", withAuth)
	{
		consignees.POST("", consignee.CreateHandler(pool))
		consignees.GET("", consignee.ListHandler(pool))
		consignees.DELETE(":consignee", consignee.DeleteHandler(pool))

		consigneeMe := consignees.Group("/me")
		{
			consigneeMe.GET("/remessas", consignment.ListByConsigneeHandler(pool))
		}
	}

	public.GET("/remessas-publicas", consignment.PublicListHandler(pool))

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

		consignments.DELETE(
			":consignment/produtos/:product",
			consignment.RemoveProductHandler(pool))
	}

	pickupLocations := t.Group("/pontos", withAuth)
	{
		pickupLocations.POST("", pickups.CreateHandler(pool))
		pickupLocations.GET("", pickups.ListHandler(pool))
		pickupLocations.PUT(":pickup", pickups.UpdateHandler(pool))
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
		clients.GET("", client.ListHandler(pool))
	}

	admin := t.Group("/admin", withAuth)
	{
		admin.GET("/saques", saque.AdminListHandler(pool))

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

		admin.POST(
			"/moderacao/saques/:id/aprovar",
			saque.AdminApproveHandler(pool),
		)

		admin.POST(
			"/moderacao/saques/:id/rejeitar",
			saque.AdminRejectHandler(pool),
		)
	}

	carts := t.Group("/carrinho")
	{
		carts.GET("/pacotes", cart.PackageOptionsHandler(pool))
		carts.GET("/dados", cart.ViewHandler(pool))
		carts.POST("/adicionar", cart.AddProductHandler(pool))
		carts.DELETE("/remover/:item", cart.RemoveHandler(pool))
		carts.PUT("/atualizar/:item", cart.QuantityHandler(pool))
	}

	payments := t.Group("/payments")
	{
		payments.POST("/session/validate",
			payment.ValidateTemporarySession)
	}

	checkouts := t.Group("/checkouts")
	{
		checkouts.POST("/confirm", checkout.ConfirmHandler(pool))
	}

	t.Static("/assets", "./public/uploads")
	t.Static("/public/uploads", "./public/uploads") // alias
}
