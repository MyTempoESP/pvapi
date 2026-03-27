package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/models/shopmodel"
)

type ConsignmentGate struct {
	Base db.Queryable
}

func (c *ConsignmentGate) Owns(u User, r Resource) (bool, error) {
	if u.Role != roles.Shop {
		return false, nil
	}

	id := r.Metadata["id"].(int)

	hasConsignment, err := shopmodel.HasConsignment(
		r.Context, c.Base,
		u.ActiveRoleID, id,
	)

	if err != nil {
		return false, err
	}

	return hasConsignment, nil
}

func (c *ConsignmentGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	if r.Method == "GET" {
		id := r.Params.ByName("consignment")
		r.Metadata["id"] = id

		if id == "" {
			return c.List(u, r)
		}
		return c.View(u, r)
	}

	if r.Method == "POST" {
		return c.Create(u, r)
	}

	if r.Method == "PUT" {
		return c.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return c.Delete(u, r)
	}

	return false, nil
}

func (c *ConsignmentGate) Create(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (c *ConsignmentGate) View(u User, r Resource) (bool, error) {
	return c.Owns(u, r)
}

func (c *ConsignmentGate) List(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (c *ConsignmentGate) Modify(u User, r Resource) (bool, error) {
	return c.Owns(u, r)
}

func (c *ConsignmentGate) Delete(u User, r Resource) (bool, error) {
	return c.Owns(u, r)
}

func NewConsignment(base db.Queryable) policymodel.Gate {
	return &ConsignmentGate{
		Base: base,
	}
}
