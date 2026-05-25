package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/policymodel"
)

type ShopGate struct {
	Base db.Queryable
}

func (s *ShopGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	id := r.Params.ByName("shop")
	r.Metadata["id"] = id

	if r.Method == "GET" {
		if r.EndPoint("/checkouts") {
			return s.Checkouts(u, r)
		}
		if r.EndPoint("/saques") {
			return s.Saques(u, r)
		}
		if id == "" {
			return s.List(u, r)
		}
		return s.View(u, r)
	}

	if r.Method == "POST" {
		if r.EndPoint("/saques") {
			return s.Saques(u, r)
		}
		return s.Create(u, r)
	}

	if r.Method == "PUT" {
		return s.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return s.Delete(u, r)
	}

	return false, nil
}

func (s *ShopGate) Create(u User, r Resource) (bool, error) {
	return u.Role == "admin", nil
}

func (s *ShopGate) View(u User, r Resource) (bool, error) {
	return u.Role == "admin", nil
}

func (s *ShopGate) List(u User, r Resource) (bool, error) {
	return u.Role == "admin", nil
}

func (s *ShopGate) Modify(u User, r Resource) (bool, error) {
	return u.Role == "admin", nil
}

func (s *ShopGate) Checkouts(u User, r Resource) (bool, error) {
	return u.Role == "shop", nil
}

func (s *ShopGate) Saques(u User, r Resource) (bool, error) {
	return u.Role == "shop", nil
}

func (s *ShopGate) Delete(u User, r Resource) (bool, error) {
	return u.Role == "admin", nil
}

func NewShop(base db.Queryable) policymodel.Gate {
	return &ShopGate{
		Base: base,
	}
}
