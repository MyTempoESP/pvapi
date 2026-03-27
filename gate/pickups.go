package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/models/shopmodel"
)

type PickupsGate struct {
	Base db.Queryable
}

func (p *PickupsGate) Owns(u User, r Resource) (bool, error) {
	if u.Role != roles.Shop {
		return false, nil
	}

	id := r.Metadata["id"]

	hasProduct, err := shopmodel.HasPickup(
		r.Context, p.Base,
		u.ActiveRoleID, id,
	)

	if err != nil {
		return false, err
	}

	return hasProduct, nil
}

func (p *PickupsGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	id := r.Params.ByName("pickup")
	r.Metadata["id"] = id

	if r.Method == "GET" {
		if id == "" {
			return p.List(u, r)
		}
		return p.View(u, r)
	}

	if r.Method == "POST" {
		return p.Create(u, r)
	}

	if r.Method == "PUT" {
		return p.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return p.Delete(u, r)
	}

	return false, nil
}

func (p *PickupsGate) Create(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (p *PickupsGate) View(u User, r Resource) (bool, error) {
	return p.Owns(u, r)
}

func (p *PickupsGate) List(u User, r Resource) (bool, error) {
	if u.Role == roles.Admin {
		return true, nil
	}

	if u.Role == roles.Shop {
		return true, nil
	}

	if u.Role == roles.Consignee {
		return true, nil
	}

	return false, nil
}

func (p *PickupsGate) Modify(u User, r Resource) (bool, error) {
	return p.Owns(u, r)
}

func (p *PickupsGate) Delete(u User, r Resource) (bool, error) {
	return p.Owns(u, r)
}

func NewPickups(base db.Queryable) policymodel.Gate {
	return &PickupsGate{
		Base: base,
	}
}
