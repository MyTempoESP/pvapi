package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/gate/roles"
)

type CategoryGate struct {
	Base db.Queryable
}

func (category *CategoryGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	id := r.Params.ByName("category")
	r.Metadata["id"] = id

	if r.Method == "POST" {
		return category.Create(u, r)
	}

	if r.Method == "GET" {
		if id == "" {
			return category.List(u, r)
		}
		return category.View(u, r)
	}

	if r.Method == "PUT" {
		return category.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return category.Delete(u, r)
	}

	return false, nil
}

func (category *CategoryGate) Create(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (category *CategoryGate) View(u User, r Resource) (bool, error) {
	// TODO: implement Owns
	return false, nil
}

func (category *CategoryGate) List(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (category *CategoryGate) Modify(u User, r Resource) (bool, error) {
	return false, nil
}

func (category *CategoryGate) Delete(u User, r Resource) (bool, error) {
	return false, nil
}

func NewCategory(base db.Queryable) policymodel.Gate {
	return &CategoryGate{
		Base: base,
	}
}
