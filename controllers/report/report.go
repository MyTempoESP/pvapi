package report

import (
    "strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"linhazero.com/pvapi/models/reportmodel"
	"linhazero.com/pvapi/util"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        reports, err := reportmodel.List(
            c.Request.Context(),
            pool,
        )

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        collection := util.ResourceCollection(reports)

        // CHANGEME
        util.Wrap(
            c, "[report]s listados com sucesso",
            collection,
        )
	}
}

func CreateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        var request reportmodel.CreateReportRequest

        err := c.ShouldBind(&request)

        if err != nil {
            util.HandleBadRequest(c, err)
            return
        }

        report := request.Prepare()

        id, err := report.Save(c.Request.Context(), pool)

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        // CHANGEME
        util.WrapLabeled(
            c, "[report] criado com sucesso",
            "id", id,
        )
	}
}

func ViewHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        reportID := c.Param("report")

        report, err := reportmodel.Find(
            c.Request.Context(),
            pool,
            reportID,
        )

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        // CHANGEME
        util.WrapLabeled(
            c, "[report] acessado com sucesso",
            "report", report.ToResource(),
        )
	}
}

func UpdateHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        reportID := c.Param("report")

        var request reportmodel.UpdateReportRequest

        err := c.ShouldBind(&request)

        if err != nil {
            util.HandleBadRequest(c, err)
            return
        }

        id, err := strconv.Atoi(reportID)

        if err != nil {
            util.HandleBadRequest(c, err)
            return
        }

        report := request.Prepare(id)

        err = report.Save(c.Request.Context(), pool)

        if err != nil {
            util.HandleGuessFromDBOperation(c, err)
            return
        }

        // CHANGEME
        util.WrapLabeled(
            c, "[report] atualizado com sucesso",
            "report", report.ToResource(),
        )
	}
}

func DeleteHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
        err := reportmodel.Delete(
            c.Request.Context(), pool,
            c.Param("report"),
        )

        if err != nil {
            util.Handle500(c, err)
            return
        }

        util.OK(
            c, "[report] deletado com sucesso",
        )
	}
}
