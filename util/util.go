package util

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/middleware/logmw"
)

func GetPgError(err error) (pge *pgconn.PgError, ok bool) {
	ok = errors.As(err, &pge)
	return
}

// aliasing for centralization
func HashPassword(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword(
		[]byte(pass), 12,
	)
}

func ComparePasswordWithHash(pass string, hashed []byte) error {
	return bcrypt.CompareHashAndPassword(hashed, []byte(pass))
}

func HandleBadRequest(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Bad request", "error", err)

	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		gin.H{
			"status":  false,
			"message": err.Error(), //FIXME:debug (error leak)
		},
	)
}

func Handle404(c *gin.Context) {
	logger := logmw.GetLogger(c)
	logger.Warn("Resource not found")

	c.AbortWithStatusJSON(
		404,
		gin.H{
			"message": "resource not found",
		},
	)
}

func Handle500(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Error("Internal server error", "error", err)

	c.AbortWithStatusJSON(
		500,
		gin.H{
			"status":  false,
			"message": err.Error(), //FIXME:debug (error leak)
		},
	)
}

// HandleGuessFromDBOperation guesses the correct error code based
// on the error returned from a database operation
func HandleGuessFromDBOperation(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		Handle404(c)
		return
	}

	Handle500(c, err)
}

type Show[R any] interface {
	ToResource() R
}

func ResourceCollection[R any, M Show[R]](collection []M) []R {
	resources := make([]R, len(collection))

	for i, item := range collection {
		resources[i] = item.ToResource()
	}

	return resources
}

func WithCollectExactlyOneRow[T any](
	ctx context.Context,
	base db.Queryable,
	query string,
	params ...any,
) (T, error) {
	// empty value to return on failure
	var zero T

	row, err := base.Query(ctx, query, params...)

	if err != nil {
		return zero, err
	}
	// called internally by pgx.CollectRows
	// defer rows.Close()

	return pgx.CollectExactlyOneRow(
		row, pgx.RowToStructByName[T],
	)
}

func WithCollect[T any](
	ctx context.Context,
	base db.Queryable,
	query string,
	params ...any,
) ([]T, error) {
	rows, err := base.Query(ctx, query, params...)

	if err != nil {
		return nil, err
	}
	// called internally by pgx.CollectRows
	// defer rows.Close()

	return pgx.CollectRows(
		rows, pgx.RowToStructByName[T],
	)
}

// WrapLabeled wraps the specified labeled resource
// into a HTTP 200 response, along with a message,
// also wraps the inner resource in a 'data' object
func WrapLabeled[R any](
	c *gin.Context,
	message string,
	label string, data R,
) {
	c.JSON(200, gin.H{
		"data": gin.H{
			"success": true,
			"message": message,
			label:     data,
		},
	})
}

// WrapLabeledFlat works just like WrapLabeled, except it
// does not wrap the inner resource inside a 'data' object,
// instead the object is provided as-is with the specified label
// func WrapLabeledFlat[R any](
// 	c *gin.Context,
// 	message string,
// 	label string, data R,
// ) {
// 	c.JSON(200, gin.H{
// 		"data": gin.H{
// 			"success": true,
// 			"message": message,
// 			label:     data,
// 		},
// 	})
// }

func WrapMap(
	c *gin.Context,
	message string,
	m gin.H,
) {
	m["success"] = true
	m["message"] = message

	c.JSON(200, gin.H{
		"data": m,
	})
}
