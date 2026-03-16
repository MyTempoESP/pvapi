package productpolicy

import (
	"github.com/gin-gonic/gin"
	"linhazero.com/pvapi/models/authmodel"
)

// Gate dispatches the method/path to the correct policy handler
func Gate(
	user *authmodel.AuthUser,
	resource string,
	params gin.Params,
) bool {
	switch resource {
	case "GET/produtos/estabelecimento":
		return ListPolicy(user)
	case "POST/produtos":
		return CreatePolicy(user)
	case "PUT/produtos/:id":
		return AlterPolicy(user, params.ByName("id"))
	}

	return false
}

func CreatePolicy(user *authmodel.AuthUser) bool {
	return user.Role == "shop"
}

func ListPolicy(user *authmodel.AuthUser) bool {
	return user.Role == "shop"
}

func AlterPolicy(
	user *authmodel.AuthUser,
	product string,
) bool {
	return user.Role == "shop" // XXX: this is only ok because on update we do an "AND shop.id = shop_id"
}
