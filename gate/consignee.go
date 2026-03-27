package gate

import (
	"strings"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/models/shopmodel"
)

type ConsigneeGate struct {
	Base db.Queryable
}

func (c *ConsigneeGate) Owns(u User, r Resource) (bool, error) {
	if u.Role != roles.Shop {
		return false, nil
	}

	id := r.Metadata["id"].(int)

	hasConsignee, err := shopmodel.HasConsignee(
		r.Context, c.Base,
		u.ActiveRoleID, id,
	)

	if err != nil {
		return false, err
	}

	return hasConsignee, nil
}

func (c *ConsigneeGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	if strings.HasPrefix(r.Path, "/monitores/me") {
		return c.Me(u, r)
	}

	if r.Method == "GET" {
		id := r.Params.ByName("consignee")
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

func (c *ConsigneeGate) Me(u User, r Resource) (bool, error) {
	return u.Role == roles.Consignee, nil
}

func (c *ConsigneeGate) Create(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (c *ConsigneeGate) View(u User, r Resource) (bool, error) {
	return c.Owns(u, r)
}

func (c *ConsigneeGate) List(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (c *ConsigneeGate) Modify(u User, r Resource) (bool, error) {
	return c.Owns(u, r)
}

func (c *ConsigneeGate) Delete(u User, r Resource) (bool, error) {
	return c.Owns(u, r)
}

func NewConsignee(base db.Queryable) policymodel.Gate {
	return &ConsigneeGate{
		Base: base,
	}
}
