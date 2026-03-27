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
	if r.Method == "GET" {
		id := r.Params.ByName("shop")
		r.Metadata["id"] = id

		if id == "" {
			return s.List(u, r)
		}
		return s.View(u, r)
	}

	if r.Method == "POST" {
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

func (s *ShopGate) Delete(u User, r Resource) (bool, error) {
	return u.Role == "admin", nil
}

func NewShop(base db.Queryable) policymodel.Gate {
	return &ShopGate{
		Base: base,
	}
}
