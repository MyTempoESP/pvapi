package apitoken

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/apitokenmodel"
	"linhazero.com/pvapi/util"
)

type CreateRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

// CreateHandler creates a new API client and returns the plaintext token once.
func CreateHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		id, plainToken, err := apitokenmodel.Create(c.Request.Context(), pool, req.Name, req.Email)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		client, err := apitokenmodel.FindByID(c.Request.Context(), pool, id)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "API client created",
			"api_client": client.ToPublic(),
			"token":   plainToken,
		})
	}
}

// ListHandler lists all API clients.
func ListHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		clients, err := apitokenmodel.List(c.Request.Context(), pool)
		if err != nil {
			util.Handle500(c, err)
			return
		}

		public := make([]apitokenmodel.APIClientPublic, len(clients))
		for i, client := range clients {
			public[i] = client.ToPublic()
		}

		c.JSON(200, gin.H{
			"success":    true,
			"api_clients": public,
		})
	}
}

// ViewHandler returns a single API client by ID.
func ViewHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		client, err := apitokenmodel.FindByID(c.Request.Context(), pool, id)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		c.JSON(200, gin.H{
			"success":    true,
			"api_client": client.ToPublic(),
		})
	}
}

// RevokeHandler revokes an API client's token.
func RevokeHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		if err := apitokenmodel.Revoke(c.Request.Context(), pool, id); err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.OK(c, "API client revoked")
	}
}
