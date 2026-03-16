package authmodel

import (
	"fmt"
	"strconv"

	jwtg "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUser struct {
	UserID string
	Role   string
}

var (
	IdentityKey = "id"
	RoleKey     = "role"
)

func GetUserFromClaims(c jwt.MapClaims) *AuthUser {
	return &AuthUser{
		UserID: c[IdentityKey].(string),
		Role:   c[RoleKey].(string),
	}
}

func GetUserFromContext(c *gin.Context) (userID int, err error) {
	var (
		claims       = jwtg.ExtractClaims(c)
		authUser     = GetUserFromClaims(claims)
		userIDString = authUser.UserID
	)

	userID, err = strconv.Atoi(userIDString)

	// wrapper for custom error instead of postgres screaming
	if err != nil {
		err = fmt.Errorf("failed to convert user id %q: %w", userIDString, err)
		return
	}

	return
}
