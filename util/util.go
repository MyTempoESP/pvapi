package util

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"

	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/middleware/logmw"
	"linhazero.com/pvapi/env"
)

var (
	ErrBadRequest         = errors.New("malformed request")
	ErrDuplicatedResource = errors.New("trying to create duplicated resource")
	ErrResourceLocked     = errors.New("resource is locked and cannot be modified")
	ErrNotFound           = errors.New("resource not found")
	ErrNotImplemented     = errors.New("function not yet implemented")
)

func NotImplemented(operation string) error {
	return fmt.Errorf(
		"could not fulfill request for operation %q: %w",
		operation, ErrNotImplemented,
	)
}

var SQL = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// Sybau 🥀😭
func Sybau(vals ...any) {}

// A is a quick alias for NamedArgs
type Args = pgx.NamedArgs

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

func Handle404(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Resource not found", "error", err)

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

func LogInfo(c *gin.Context, message string, a ...any) {
	logger := logmw.GetLogger(c)
	logger.Info(message, a...)
}

func LogWarn(c *gin.Context, message string, a ...any) {
	logger := logmw.GetLogger(c)
	logger.Warn(message, a...)
}

func LogError(c *gin.Context, message string, a ...any) {
	logger := logmw.GetLogger(c)
	logger.Error(message, a...)
}

// HandleGuess guesses the correct error code based
// on the error returned from a generic operation
func HandleGuess(c *gin.Context, err error) {
	if errors.Is(err, ErrBadRequest) {
		HandleBadRequest(c, err)
		return
	}

	if errors.Is(err, ErrNotFound) {
		Handle404(c, err)
		return
	}

	Handle500(c, err)
}

// HandleGuessFromDBOperation guesses the correct error code based
// on the error returned from a database operation
func HandleGuessFromDBOperation(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		Handle404(c, err)
		return
	}

	if errors.Is(err, ErrBadRequest) {
		HandleBadRequest(c, err)
		return
	}

	if errors.Is(err, ErrDuplicatedResource) {
		HandleDuplicatedResource(c, err)
		return
	}

	if errors.Is(err, ErrResourceLocked) {
		HandleResourceLocked(c, err)
		return
	}

	if errors.Is(err, ErrNotFound) {
		Handle404(c, err)
		return
	}

	Handle500(c, err)
}

// HandleDuplicatedResource handles ErrDuplicatedResource errors with a
// machine-readable 409 Conflict response.
func HandleGone(c *gin.Context, message string) {
	c.AbortWithStatusJSON(
		410,
		gin.H{
			"message": message,
		},
	)
}

func HandleDuplicatedResource(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Duplicated resource", "error", err)

	c.AbortWithStatusJSON(
		http.StatusConflict,
		gin.H{
			"success": false,
			"message": "resource already exists",
			"code":    "duplicated_resource",
		},
	)
}

func HandleResourceLocked(c *gin.Context, err error) {
	logger := logmw.GetLogger(c)
	logger.Warn("Resource locked", "error", err)

	c.AbortWithStatusJSON(
		http.StatusConflict,
		gin.H{
			"success": false,
			"message": "não é possível remover: produto possui reservas ativas em checkouts",
			"code":    "resource_locked",
		},
	)
}

type Show[R any] interface {
	ToResource() R
}

func ToResourceCollection[M any, R any](
	collection []M,
	toResource func(M) R,
) []R {
	resources := make([]R, len(collection))
	for i, item := range collection {
		resources[i] = toResource(item)
	}
	return resources
}

func ResourceCollection[R any, M Show[R]](collection []M) []R {
	return ToResourceCollection(collection, M.ToResource)
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
// into a HTTP 200 response, along with a message.
func WrapLabeled[R any](
	c *gin.Context,
	message string,
	label string, data R,
) {
	c.JSON(200, gin.H{
		"success": true,
		"message": message,
		label:     data,
	})
}

func Wrap[R any](
	c *gin.Context,
	message string,
	data R,
) {
	WrapLabeled(c, message, "data", data)
}

func WrapMap(
	c *gin.Context,
	message string,
	json gin.H,
) {
	json["success"] = true
	json["message"] = message
	c.JSON(200, json)
}

func OK(
	c *gin.Context,
	message string,
) {
	c.JSON(200, gin.H{
		"success": true,
		"message": message,
	})
}

func Transaction(
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, db.Queryable) error,
) error {
	return pgx.BeginFunc(ctx, pool,
		func(tx pgx.Tx) error {
			return f(ctx, tx)
		},
	)
}

func TransactionAware(
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, pgx.Tx) error,
) error {
	return pgx.BeginFunc(ctx, pool,
		func(tx pgx.Tx) error {
			return f(ctx, tx)
		},
	)
}

func TransactionReturn[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, db.Queryable) (T, error),
) (T, error) {
	var zero T

	tx, err := pool.Begin(ctx)

	if err != nil {
		return zero, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	t, err := f(ctx, tx)

	if err != nil {
		return zero, err
	}

	err = tx.Commit(ctx)

	return t, err
}

func TransactionAwareReturn[T any](
	ctx context.Context,
	pool *pgxpool.Pool,
	f func(context.Context, pgx.Tx) (T, error),
) (T, error) {
	var zero T

	tx, err := pool.Begin(ctx)

	if err != nil {
		return zero, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	t, err := f(ctx, tx)

	if err != nil {
		return zero, err
	}

	err = tx.Commit(ctx)

	return t, err
}

func Merge(args ...Args) (merge Args) {
	merge = make(Args)
	for _, a := range args {
		maps.Copy(merge, a)
	}
	return
}

func Coalesce[T any](nullable *T, fallback T) T {
	if nullable != nil {
		return *nullable
	}
	return fallback
}

// Bake exec's a *pgx.Batch, based on its Len().
// Every query in the batch gets executed or returns
// early on error.
func Bake(
	ctx context.Context,
	base db.Queryable,
	batch *pgx.Batch,
) (err error) {
	b := base.SendBatch(ctx, batch)
	defer func() { _ = b.Close() }()

	for range batch.Len() {
		_, err = b.Exec()

		if err != nil {
			return
		}
	}

	err = b.Close()

	return
}

func GetDomainURL() (url string, err error) {
	domain, err := env.DomainName()

	if err != nil {
		return
	}

	protocol := "http"

	if env.HTTPSEnabled {
		protocol = "https"
	}

	url = fmt.Sprintf(
		"%s://%s",
		protocol, domain,
	)

	return
}

func GetImageURL(base string) (url string, err error) {
	domain, err := GetDomainURL()

	if err != nil {
		return
	}

	url = fmt.Sprintf(
		"%s/public/uploads/%s",
		domain, base,
	)

	return
}

