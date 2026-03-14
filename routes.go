package main

import (
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/auth"
	"linhazero.com/pvapi/user"
)

func setupRoutes(
	t *gin.Engine,
	authMiddleware *jwt.GinJWTMiddleware,
	pool *pgxpool.Pool,
) {
	auth.RegisterAuthRoutes(t, authMiddleware, pool)

	loggedIn := t.Group("", authMiddleware.MiddlewareFunc())
	{
		loggedIn.GET("/logout", authMiddleware.LogoutHandler)
		loggedIn.GET("/user", user.UserDataHandler(pool))
	}

	//admin := t.Group("/admin", auth.MiddlewareFunc())
	{
	}
}
