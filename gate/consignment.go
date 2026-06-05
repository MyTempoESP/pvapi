package gate

import (
	"errors"
	"strconv"

	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/consigneemodel"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/models/shopmodel"
)

type ConsignmentGate struct {
	Base db.Queryable
}

func (c *ConsignmentGate) consigneeOwns(
	u User,
	r Resource,
	id any,
) (bool, error) {
	idStr, ok := id.(string)

	if !ok {
		return false, errors.New("search query is not a string")
	}

	consignmentID, err := strconv.Atoi(idStr)
	if err != nil {
		return false, err
	}

	hasConsignment, err := consigneemodel.HasConsignment(
		r.Context, c.Base,
		u.ActiveRoleID, consignmentID,
	)

	if err != nil {
		return false, err
	}

	return hasConsignment, nil
}

func (c *ConsignmentGate) shopOwns(
	u User,
	r Resource,
	id any,
) (bool, error) {
	idStr, ok := id.(string)

	if !ok {
		return false, errors.New("search query is not a string")
	}

	consignmentID, err := strconv.Atoi(idStr)
	if err != nil {
		return false, err
	}

	hasConsignment, err := shopmodel.HasConsignment(
		r.Context, c.Base,
		u.ActiveRoleID, consignmentID,
	)

	if err != nil {
		return false, err
	}

	return hasConsignment, nil
}

func (c *ConsignmentGate) Owns(u User, r Resource) (bool, error) {
	id := r.Metadata["id"]

	if u.Role != roles.Shop {
		return false, nil
	}

	return c.shopOwns(u, r, id)
}

func (c *ConsignmentGate) InvolvedIn(u User, r Resource) (bool, error) {
	id := r.Metadata["id"]

	if u.Role == roles.Consignee {
		return c.consigneeOwns(u, r, id)
	}

	return c.shopOwns(u, r, id)
}

func (c *ConsignmentGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	id := r.Params.ByName("consignment")
	r.Metadata["id"] = id

	if r.EndPoint("/produtos") {
		if r.Method == "GET" {
			return c.View(u, r)
		}

		if r.Method == "POST" {
			return c.Modify(u, r)
		}

		return false, nil
	}

	if r.EndPoint("/produtos/:product") {
		if r.Method == "PUT" {
			return c.Modify(u, r)
		}

		if r.Method == "DELETE" {
			return c.Modify(u, r)
		}

		return false, nil
	}

	if r.EndPoint("/pontos") {
		if r.Method == "POST" {
			return c.Modify(u, r)
		}
		return false, nil
	}

	if r.EndPoint("/pontos/:pickup") {
		if r.Method == "DELETE" {
			return c.Modify(u, r)
		}
		return false, nil
	}

	if r.Method == "GET" {
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
	return c.InvolvedIn(u, r)
}

func (c *ConsignmentGate) List(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop || u.Role == roles.Consignee, nil
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
