package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/policymodel"
	"linhazero.com/pvapi/models/shopmodel"
)

type ProductGate struct {
	Base db.Queryable
}

func (p *ProductGate) Owns(u User, r Resource) (bool, error) {
	if u.Role != roles.Shop {
		return false, nil
	}

	id := r.Metadata["id"]

	hasProduct, err := shopmodel.HasProduct(
		r.Context, p.Base,
		u.ActiveRoleID, id,
	)

	if err != nil {
		return false, err
	}

	return hasProduct, nil
}

func (p *ProductGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	id := r.Params.ByName("product")
	r.Metadata["id"] = id

	if r.Path == "/produtos/estabelecimento" {
		return p.List(u, r)
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

func (p *ProductGate) Create(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (p *ProductGate) View(u User, r Resource) (bool, error) {
	return p.Owns(u, r)
}

func (p *ProductGate) List(u User, r Resource) (bool, error) {
	// admin can't list his products because he does not have any
	return u.Role == roles.Shop, nil
}

func (p *ProductGate) Modify(u User, r Resource) (bool, error) {
	return p.Owns(u, r)
}

func (p *ProductGate) Delete(u User, r Resource) (bool, error) {
	return p.Owns(u, r)
}

func NewProduct(base db.Queryable) policymodel.Gate {
	return &ProductGate{
		Base: base,
	}
}
