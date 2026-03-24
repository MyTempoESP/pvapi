package consignmentpolicy

import (
	"github.com/gin-gonic/gin"
	"linhazero.com/pvapi/models/authmodel"
)

// this package should include some db shit

// Gate dispatches the method/path to the correct policy handler
func Gate(
	user *authmodel.AuthUser,
	resource string,
	params gin.Params,
) bool {
	if user.Role == "admin" {
		return true
	}

	switch resource {
	case "POST/remessas":
		return CreatePolicy(user)
	case "GET/remessas":
		return ListPolicy(user)
	}

	return false
}

func CreatePolicy(user *authmodel.AuthUser) bool {
	return user.Role == "shop"
}

func ListPolicy(user *authmodel.AuthUser) bool {
	return user.Role == "shop"
}
