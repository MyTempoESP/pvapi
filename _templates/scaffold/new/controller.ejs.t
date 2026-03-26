---
to: controllers/<%= h.changeCase.no(name) %>/<%= h.changeCase.no(name) %>.go
---
<%
    Flat = h.changeCase.no(name)
    collection = h.inflection.pluralize(Flat)

    Class = h.changeCase.pascal(name)
    Model = Flat + "model"

    Char = name.charAt(0)
%>package <%= Flat %>

import (
    "strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/<%= Model %>"
	"linhazero.com/pvapi/util"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        <%= collection %>, err := <%= Model %>.List(
            c.Request.Context(),
            pool,
        )

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        collection := util.ResourceCollection(<%= collection %>)

        // CHANGEME
        util.Wrap(
            c, "[<%= Flat %>]s listados com sucesso",
            collection,
        )
	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        var request <%= Model %>.Create<%= Class %>Request

        err := c.ShouldBind(&request)

        if err != nil {
            util.HandleBadRequest(c, err)
            return
        }

        <%= Flat %> := request.Prepare()

        id, err := <%= Flat %>.Save(c.Request.Context(), pool)

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        // CHANGEME
        util.WrapLabeled(
            c, "[<%= Flat %>] criado com sucesso",
            "id", id,
        )
	}
}

func ViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        <%= Flat %>ID := c.Param("<%= Flat %>")

        <%= Flat %>, err := <%= Model %>.Find(
            c.Request.Context(),
            pool,
            <%= Flat %>ID,
        )

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        // CHANGEME
        util.WrapLabeled(
            c, "[<%= Flat %>] acessado com sucesso",
            "<%= Flat %>", <%= Flat %>.ToResource(),
        )
	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        <%= Flat %>ID := c.Param("<%= Flat %>")

        var request <%= Model %>.Update<%= Class %>Request

        err := c.ShouldBind(&request)

        if err != nil {
            util.HandleBadRequest(c, err)
            return
        }

        id, err := strconv.Atoi(<%= Flat %>ID)

        if err != nil {
            util.HandleBadRequest(c, err)
            return
        }

        <%= Flat %> := request.Prepare(id)

        err = <%= Flat %>.Save(c.Request.Context(), pool)

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        // CHANGEME
        util.WrapLabeled(
            c, "[<%= Flat %>] atualizado com sucesso",
            "<%= Flat %>", <%= Flat %>.ToResource(),
        )
	}
}

func DeleteHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        err := <%= Model %>.Delete(
            c.Request.Context(), pool,
            c.Param("<%= Flat %>"),
        )

        if err != nil {
            util.Handle500(c, err)
            return
        }

        util.OK(
            c, "[<%= Flat %>] deletado com sucesso",
        )
	}
}
