package report

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/consigneemodel"
	"linhazero.com/pvapi/models/reportmodel"
	"linhazero.com/pvapi/models/shopmodel"
	"linhazero.com/pvapi/util"
)

func ListHandler(
	pool *pgxpool.Pool,
) func(c *gin.Context) {
	return func(c *gin.Context) {
		user := authmodel.GetUserFromContext(c)

		var (
			reports []reportmodel.Report
			err error
		)

		switch user.Role {
		case roles.Shop:
			reports, err = reportmodel.ListByShop(
				c.Request.Context(), pool,
				user.ActiveRoleID,
			)
		case roles.Consignee:
			reports, err = reportmodel.ListByConsignee(
				c.Request.Context(), pool,
				user.ActiveRoleID,
			)
		default:
			err = errors.New(
				"user not authorized (gate failure?)")
		}

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

		user := authmodel.GetUserFromContext(c)

		// TODO: frontend sends occurrenceScope="shipment" instead of "consignment"
		if request.Scope == "shipment" {
			request.Scope = "consignment"
		}

		err = validateScope(c, pool, user, request.Scope, int(request.OrderID))
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
		reportID, err := strconv.Atoi(c.Param("report"))

		if err != nil {
			util.Handle500(c, err)
			return
		}

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
		reportID, err := strconv.Atoi(c.Param("report"))

		if err != nil {
			util.Handle500(c, err)
			return
		}

		var request reportmodel.UpdateReportRequest

		err = c.ShouldBind(&request)

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		blueprint := request.Prepare(reportID)

		err = blueprint.Save(c.Request.Context(), pool)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		updated, err := reportmodel.Find(c.Request.Context(), pool, reportID)
		if err != nil {
			util.HandleGuessFromDBOperation(c, err)
			return
		}

		util.WrapLabeled(
			c, "[report] atualizado com sucesso",
			"report", updated.ToResource(),
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

func validateScope(
	c *gin.Context,
	pool *pgxpool.Pool,
	user authmodel.AuthUser,
	scope string,
	entityID int,
) error {
	ctx := c.Request.Context()

	switch scope {
	case "consignment":
		return validateConsignmentScope(ctx, pool, user, entityID)
	case "product":
		return validateProductScope(ctx, pool, user, entityID)
	default:
		return fmt.Errorf("unknown report scope: %q", scope)
	}
}

func validateConsignmentScope(
	ctx context.Context,
	pool *pgxpool.Pool,
	user authmodel.AuthUser,
	consignmentID int,
) error {
	var (
		hasResource bool
		err         error
	)

	switch user.Role {
	case roles.Shop:
		hasResource, err = shopmodel.HasConsignment(ctx, pool, user.ActiveRoleID, consignmentID)
	case roles.Consignee:
		hasResource, err = consigneemodel.HasConsignment(ctx, pool, user.ActiveRoleID, consignmentID)
	default:
		return errors.New("user not authorized to create reports")
	}

	if err != nil {
		return fmt.Errorf("failed to verify consignment ownership: %w", err)
	}

	if !hasResource {
		return errors.New("consignment not found or does not belong to this shop")
	}

	return nil
}

func validateProductScope(
	ctx context.Context,
	pool *pgxpool.Pool,
	user authmodel.AuthUser,
	productID int,
) error {
	var (
		hasResource bool
		err         error
	)

	switch user.Role {
	case roles.Shop:
		hasResource, err = shopmodel.HasProduct(ctx, pool, user.ActiveRoleID, productID)
	case roles.Consignee:
		hasResource, err = consigneemodel.HasProduct(ctx, pool, user.ActiveRoleID, productID)
	default:
		return errors.New("user not authorized to create reports")
	}

	if err != nil {
		return fmt.Errorf("failed to verify product ownership: %w", err)
	}

	if !hasResource {
		return errors.New("product not found or does not belong to this shop")
	}

	return nil
}
