package auth

import (
	"net/http"
	"time"

	jwtg "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"linhazero.com/pvapi/controllers/policy"
	"linhazero.com/pvapi/env"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/util"
	"linhazero.com/pvapi/util/userutils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type authFn func(c *gin.Context) (any, error)

type login struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterAuthRoutes(
	t *gin.Engine,
	handle *jwtg.GinJWTMiddleware,
	pool *pgxpool.Pool,
) {
	t.POST("/login", handle.LoginHandler)
	t.POST("/refresh", handle.RefreshHandler)
	t.POST("/register", authRegistrator(pool))
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

	bouncer := policy.New(pool)

	return &jwtg.GinJWTMiddleware{
		Realm:       "linhazero",
		Key:         []byte(secret),
		Timeout:     time.Hour * 24,
		MaxRefresh:  time.Hour * 24 * 7,
		IdentityKey: authmodel.IdentityKey,
		PayloadFunc: withUser,

		SendCookie:     true,
		SecureCookie:   secureCookie,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,

		IdentityHandler: retrieveUser,
		Authenticator:   authenticator(pool),
		Authorizer:      authorizationHandler(bouncer),

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

func authorizationHandler(
	bouncer *policymodel.Bouncer,
) func(c *gin.Context, data any) bool {
	return func(c *gin.Context, data any) bool {
		authUser, ok := data.(*authmodel.AuthUser)

		if !ok {
			return false
		}

		return bouncer.Dispatch(c, authUser)
	}
}

func authenticator(pool *pgxpool.Pool) authFn {
	return func(c *gin.Context) (any, error) {
		var l login

		err := c.ShouldBind(&l)

		if err != nil {
			return nil, jwtg.ErrMissingLoginValues
		}

		userID, err := usermodel.Login(
			c.Request.Context(), pool,
			l.Email,
			l.Password,
		)

		if err != nil {
			util.LogError(
				c, "Error logging in user",
				"loginForm", l,
				"error", err,
			)
			return nil, jwtg.ErrFailedAuthentication
		}

		permissions, err := usermodel.GetUserPermissions(
			c.Request.Context(),
			pool,
			userID,
		)

		if err != nil {
			util.LogError(
				c, "Error grabbing user permissions",
				"error", err,
			)
			return nil, jwtg.ErrFailedTokenCreation
		}

		return &authmodel.AuthUser{
			UserID:       userID,
			Role:         permissions.Highest,
			ActiveRoleID: permissions.ActiveRoleID,
		}, nil
	}
}

func withUser(data any) jwt.MapClaims {
	v, ok := data.(*authmodel.AuthUser)

	if ok {
		return jwt.MapClaims{
			authmodel.IdentityKey: v.UserID,
			authmodel.RoleKey:     v.Role,
			authmodel.RoleIDKey:   v.ActiveRoleID,
		}
	}

	return jwt.MapClaims{}
}

func retrieveUser(c *gin.Context) any {
	claims := jwtg.ExtractClaims(c)
	return authmodel.GetUserFromClaims(claims)
}

func authRegistrator(pool *pgxpool.Pool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var r usermodel.UserRegisterRequest

		err := c.ShouldBind(&r)

		// FIXME/TODO: check pre-existing user

		if err != nil {
			util.HandleBadRequest(c, err)
			return
		}

		user, err := userutils.NewUser(
			r.Name,
			r.Email,
			r.Password,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		_, err = util.TransactionReturn[int](
			c.Request.Context(),
			pool,
			user.MustSave,
		)

		if err != nil {
			util.Handle500(c, err)
			return
		}

		c.JSON(200, gin.H{"message": "Successfully registered"})
	}
}
