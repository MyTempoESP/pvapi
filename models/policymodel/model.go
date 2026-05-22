package policymodel

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/util"
)

type Resource struct {
	Path     string
	Method   string
	Params   gin.Params
	Context  context.Context
	Metadata map[string]any
}

func NewResource(c *gin.Context) Resource {
	return Resource{
		Path:     c.FullPath(),
		Method:   c.Request.Method,
		Params:   c.Params,
		Context:  c.Request.Context(),
		Metadata: make(map[string]any),
	}
}

// EndPointPrefix returns true if r.Path
// is prefixed by 'path'
func (r Resource) EndPointPrefix(path string) bool {
	return strings.HasPrefix(r.Path, path)
}

// EndPoint returns true if r.Path ends in path
func (r Resource) EndPoint(path string) bool {
	return strings.HasSuffix(r.Path, path)
}

type User *authmodel.AuthUser
type Gate interface {
	Dispatch(User, Resource) (bool, error)
}

type Guard struct {
	Path string
	Gate
}

type Bouncer struct {
	Base      db.Queryable
	Protected []Guard
}

func NewBouncer(base db.Queryable) *Bouncer {
	return &Bouncer{
		Base:      base,
		Protected: make([]Guard, 0),
	}
}

func (b *Bouncer) Register(g Guard) {
	b.Protected = append(b.Protected, g)
}

func (b *Bouncer) Dispatch(c *gin.Context, user User) bool {
	resource := NewResource(c)

	for _, g := range b.Protected {
		if strings.HasPrefix(resource.Path, g.Path) {
			isAllowed, err := g.Dispatch(user, resource)

			if err != nil {
				util.LogWarn(
					c, "Authorization error",
					"metadata", resource.Metadata,
					"gate", g.Path,
					"error", err,
				)
				// block access just to be safe
				return false
			}

			if !isAllowed {
				util.LogInfo(
					c, "Unauthorized",
					"metadata", resource.Metadata,
					"gate", g.Path,
					"user", user,
					"resource", resource.Path,
				)
			}

			return isAllowed
		}
	}

	// default simple auth
	switch resource.Path {
	case "/logout":
		fallthrough
	case "/level":
		fallthrough
	case "/user":
		return true

	case "/perms/admin":
		return user.Role == "admin"
	case "/perms/monitor":
		return user.Role == "consignee"
	case "/perms/shop":
		return user.Role == "shop"
	}

	return false
}
