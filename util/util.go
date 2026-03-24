package util

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
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
	resources := make([]R, 0, len(collection))

	for _, item := range collection {
		resources = append(resources, item.ToResource())
	}

	return resources
}
