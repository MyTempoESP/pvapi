package shoppolicy

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
	// Rota publica
	case "GET/estabelecimentos":
		return ListPolicy(user)
	case "POST/estabelecimentos":
		return CreatePolicy(user)
	case "GET/estabelecimentos/:shop":
		return ViewPolicy(user, params.ByName("shop"))
	case "PUT/estabelecimentos/:shop":
		return AlterPolicy(user, params.ByName("shop"))
	}

	return false
}

func CreatePolicy(user *authmodel.AuthUser) bool {
	return false // admin only
}

func ListPolicy(user *authmodel.AuthUser) bool {
	return false // admin only
}

func AlterPolicy(
	user *authmodel.AuthUser,
	shop string,
) bool {
	return user.Role == "shop"
}

func ViewPolicy(
	user *authmodel.AuthUser,
	shop string,
) bool {
	return user.Role == "shop"
}
