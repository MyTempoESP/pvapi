package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	jwtg "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"linhazero.com/pvapi/env"
	"linhazero.com/pvapi/query"
	"linhazero.com/pvapi/util"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"golang.org/x/crypto/bcrypt"
)

type authFn func(c *gin.Context) (any, error)

type login struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type register struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthUser struct {
	UserID string
}

var (
	identityKey = "id"
)

func RegisterAuthRoutes(
	t *gin.Engine,
	handle *jwtg.GinJWTMiddleware,
	pool *pgxpool.Pool,
) {
	t.POST("/login", handle.LoginHandler)
	t.POST("/refresh", handle.RefreshHandler)
	t.POST("/register", authRegistrator(pool))
}

func GetUserFromClaims(c jwt.MapClaims) *AuthUser {
	return &AuthUser{
		UserID: c[identityKey].(string),
	}
}

func GenerateJwtInitParams(
	pool *pgxpool.Pool,
) (*jwtg.GinJWTMiddleware, error) {
	// whether or not to enable secure cookies
	secureCookie := false
	secureCookieSetting, err := env.SecureCookies()

	if err != nil {
		return nil, err
	}

	if secureCookieSetting == "yes" {
		secureCookie = true
	}

	secret, err := env.JwtSecretKey()

	if err != nil {
		return nil, err
	}

	return &jwtg.GinJWTMiddleware{
		Realm:       "linhazero",
		Key:         []byte(secret),
		Timeout:     time.Minute * 15,
		MaxRefresh:  time.Hour * 24 * 7,
		IdentityKey: identityKey,
		PayloadFunc: withUser,

		SendCookie:     true,
		SecureCookie:   secureCookie,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,

		IdentityHandler: retrieveUser,
		Authenticator:   authenticator(pool),
		Authorizer:      authorize,

		Unauthorized:   unauthorized,
		LogoutResponse: logoutResponse,
		TokenLookup:    "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:  "Bearer",
		TimeFunc:       time.Now,
	}, nil
}

func logoutResponse(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Succesfully logged out",
	})
}

func unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

func authorize(c *gin.Context, data any) bool {
	_, ok := data.(*AuthUser)
	return ok
}

func authenticator(pool *pgxpool.Pool) authFn {
	return func(c *gin.Context) (any, error) {
		var l login

		err := c.ShouldBind(&l)

		if err != nil {
			return nil, jwtg.ErrMissingLoginValues
		}

		userMail := l.Email
		password := l.Password

		var hashed string
		var userID int

		err = pool.QueryRow(
			c.Request.Context(),
			query.UserPasswordQuery, userMail,
		).Scan(&userID, &hashed)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, jwtg.ErrFailedAuthentication
			}

			return nil, err
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(hashed), []byte(password),
		)

		if err != nil {
			return nil, jwtg.ErrFailedAuthentication
		}

		stringUserID := strconv.Itoa(userID)

		return &AuthUser{
			UserID: stringUserID,
		}, nil
	}
}

func withUser(data any) jwt.MapClaims {
	v, ok := data.(*AuthUser)

	if ok {
		// map identity key to username
		return jwt.MapClaims{
			identityKey: v.UserID,
		}
	}

	return jwt.MapClaims{}
}

func retrieveUser(c *gin.Context) any {
	claims := jwtg.ExtractClaims(c)
	return GetUserFromClaims(claims)
}

func authRegistrator(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var r register

		err := c.ShouldBind(&r)

		// FIXME/TODO: check pre-existing user

		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{
					"status":  false,
					"message": err.Error(), //FIXME:debug (error leak)
				})
			return
		}

		var (
			userName     = r.Name
			userEmail    = r.Email
			userPassword = r.Password
		)

		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(userPassword), 12,
		)

		// TODO: Check error type, password too long etc.
		if err != nil {
			util.AbortWithInternalError(c, err)
			return
		}

		// rollback!
		err = pgx.BeginFunc(
			context.Background(),
			pool,
			func(tx pgx.Tx) error {
				_, err := tx.Exec(
					context.Background(),
					query.UserRegisterQuery,
					userName, userEmail, passwordHash,
				)
				return err
			},
		)

		if err != nil {
			util.AbortWithInternalError(c, err)
			return
		}

		c.JSON(200, gin.H{"message": "Successfully registered"})
	}
}
