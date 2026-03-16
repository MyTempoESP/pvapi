package shop

import "linhazero.com/pvapi/authmodel"

func ViewPolicy(
	user *authmodel.AuthUser,
	shop string,
) bool {
	return user.Role == "shop"
}
