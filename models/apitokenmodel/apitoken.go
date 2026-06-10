package apitokenmodel

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/model"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"
)

const (
	TokenPrefix = "pv_"
	TokenBytes  = 32 // 256-bit token
)

var (
	ErrTokenRevoked      = errors.New("token is revoked")
	ErrTokenNotFound     = errors.New("token not found")
	ErrAPIClientNotFound = errors.New("API client not found")
)

type APIClientRow struct {
	ID        int        `db:"id"`
	Name      string     `db:"name"`
	Email     string     `db:"email"`
	TokenHash string     `db:"token_hash"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

type APIClientPublic struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

func (r APIClientRow) ToPublic() APIClientPublic {
	return APIClientPublic{
		ID:        r.ID,
		Name:      r.Name,
		Email:     r.Email,
		CreatedAt: r.CreatedAt,
		RevokedAt: r.RevokedAt,
	}
}

// TokenValidResult holds the result of token validation.
type TokenValidResult struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
	RevokedAt *time.Time
}

// GenerateToken creates a new random token with the "pv_" prefix.
func GenerateToken() (string, error) {
	buf := make([]byte, TokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return TokenPrefix + hex.EncodeToString(buf), nil
}

// HashToken returns the SHA-256 hex digest of the token.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Create inserts a new API client with a generated token.
// Returns the plaintext token and the client ID.
func Create(
	ctx context.Context,
	base db.Queryable,
	name, email string,
) (id int, plainToken string, err error) {
	plainToken, err = GenerateToken()
	if err != nil {
		return 0, "", err
	}

	tokenHash := HashToken(plainToken)

	id, err = model.Create(ctx, base, query.CreateAPIClientQuery, name, email, tokenHash)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create API client: %w", err)
	}

	return id, plainToken, nil
}

func List(ctx context.Context, base db.Queryable) ([]APIClientRow, error) {
	return util.WithCollect[APIClientRow](ctx, base, query.ListAPIClientsQuery)
}

func FindByID(ctx context.Context, base db.Queryable, id int) (APIClientRow, error) {
	return util.WithCollectExactlyOneRow[APIClientRow](
		ctx, base, query.FindAPIClientByIDQuery, id,
	)
}

// ValidateToken looks up the token hash in the DB.
// Returns ErrTokenNotFound if the hash doesn't exist,
// ErrTokenRevoked if the client is revoked.
func ValidateToken(
	ctx context.Context,
	base db.Queryable,
	plainToken string,
) (TokenValidResult, error) {
	tokenHash := HashToken(plainToken)

	row, err := util.WithCollectExactlyOneRow[APIClientRow](
		ctx, base, query.FindAPIClientByTokenHashQuery, tokenHash,
	)
	if err != nil {
		return TokenValidResult{}, ErrTokenNotFound
	}

	if row.RevokedAt != nil {
		return TokenValidResult{}, ErrTokenRevoked
	}

	return TokenValidResult{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
		RevokedAt: row.RevokedAt,
	}, nil
}

// Revoke sets revoked_at to now for the given API client.
// Returns ErrAPIClientNotFound if no active row is found.
func Revoke(ctx context.Context, base db.Queryable, id int) error {
	ct, err := base.Exec(ctx, query.RevokeAPIClientQuery, id)
	if err != nil {
		return fmt.Errorf("failed to revoke API client: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrAPIClientNotFound
	}
	return nil
}
