package authmodel

import (
	jwtg "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUser struct {
	UserID       int
	Role         string
	ActiveRoleID int
}

var (
	IdentityKey = "id"
	RoleKey     = "role"
	RoleIDKey   = "rid"
)

func GetUserFromClaims(c jwt.MapClaims) *AuthUser {
	return &AuthUser{
		UserID:       int(c[IdentityKey].(float64)),
		Role:         c[RoleKey].(string),
		ActiveRoleID: int(c[RoleIDKey].(float64)),
	}
}

func FetchUser(c *gin.Context) (AuthUser, bool) {
	claims := jwtg.ExtractClaims(c)

	_, ok := claims[IdentityKey]

	if !ok {
		return AuthUser{}, false
	}
	
	authUser := GetUserFromClaims(claims)

	return *authUser, true
}

func GetUserFromContext(c *gin.Context) AuthUser {
	var (
		claims   = jwtg.ExtractClaims(c)
		authUser = GetUserFromClaims(claims)
	)

	return *authUser
}

