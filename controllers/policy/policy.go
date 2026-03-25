package policy

import (
	"strings"

	"github.com/gin-gonic/gin"
	"linhazero.com/pvapi/models/authmodel"
	"linhazero.com/pvapi/policies/consigneepolicy"
	"linhazero.com/pvapi/policies/consignmentpolicy"
	"linhazero.com/pvapi/policies/pickupspolicy"
	"linhazero.com/pvapi/policies/productpolicy"
	"linhazero.com/pvapi/policies/shoppolicy"
)

func Gate(c *gin.Context, authUser *authmodel.AuthUser) bool {
	path := c.FullPath()
	method := c.Request.Method

	resource := method + path

	if strings.HasPrefix(path, "/estabelecimentos") {
		return shoppolicy.Gate(authUser, resource, c.Params)
	}

	if strings.HasPrefix(path, "/produtos") {
		return productpolicy.Gate(authUser, resource, c.Params)
	}

	if strings.HasPrefix(path, "/monitores") {
		return consigneepolicy.Gate(authUser, resource, c.Params)
	}

	if strings.HasPrefix(path, "/remessas") {
		return consignmentpolicy.Gate(authUser, resource, c.Params)
	}

	if strings.HasPrefix(path, "/pontos") {
		return pickupspolicy.Gate(authUser, resource, c.Params)
	}

	switch path {
	case "/logout":
		fallthrough
	case "/level":
		fallthrough
	case "/user":
		return true

	case "/perms/admin":
		return authUser.Role == "admin"
	case "/perms/monitor":
		return authUser.Role == "consignee"
	case "/perms/shop":
		return authUser.Role == "shop"
	}

	return false
}
