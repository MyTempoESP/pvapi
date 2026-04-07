package client

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/clientmodel"
	"linhazero.com/pvapi/util"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		clients, err := clientmodel.List(
			c.Request.Context(),
			pool,
			user.ActiveRoleID, // shop
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		collection := util.ResourceCollection(clients)

		// CHANGEME
		util.Wrap(
			c, "[client]s listados com sucesso",
			collection,
		)
	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request clientmodel.CreateClientRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		client := request.Prepare()

		id, err := client.Save(c.Request.Context(), pool)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		// CHANGEME
		util.WrapLabeled(
			c, "[client] criado com sucesso",
			"id", id,
		)
	}
}

func ViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		clientID := c.Param("client")

		client, err := clientmodel.Find(
			c.Request.Context(),
			pool,
			clientID,
		)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		// CHANGEME
		util.WrapLabeled(
			c, "[client] acessado com sucesso",
			"client", client.ToResource(),
		)
	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		clientID := c.Param("client")

		var request clientmodel.UpdateClientRequest

		err := c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		id, err := strconv.Atoi(clientID)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		client := request.Prepare(id)

		err = client.Save(c.Request.Context(), pool)

		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		// CHANGEME
		util.WrapLabeled(
			c, "[client] atualizado com sucesso",
			"client", client.ToResource(),
		)
	}
}

func DeleteHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := clientmodel.Delete(
			c.Request.Context(), pool,
			c.Param("client"),
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		util.OK(
			c, "[client] deletado com sucesso",
		)
	}
}
