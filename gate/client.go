package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/policymodel"
)

type ClientGate struct {
	Base db.Queryable
}

func (client *ClientGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	// Try both param names: /clientes/:client and /clients/:id
	id := r.Params.ByName("client")
	if id == "" {
		id = r.Params.ByName("id")
	}
	r.Metadata["id"] = id

	if r.Method == "GET" {
		if r.EndPoint("/sales") {
			return client.Sales(u, r)
		}
		if id == "" {
			return client.List(u, r)
		}
		return client.View(u, r)
	}

	if r.Method == "PUT" {
		return client.Modify(u, r)
	}

	if r.Method == "DELETE" {
		return client.Delete(u, r)
	}

	return false, nil
}

func (client *ClientGate) Create(u User, r Resource) (bool, error) {
	return false, nil
}

func (client *ClientGate) View(u User, r Resource) (bool, error) {
	return false, nil
}

func (client *ClientGate) List(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (client *ClientGate) Modify(u User, r Resource) (bool, error) {
	return false, nil
}

func (client *ClientGate) Sales(u User, r Resource) (bool, error) {
	return u.Role == roles.Shop, nil
}

func (client *ClientGate) Delete(u User, r Resource) (bool, error) {
	return false, nil
}

func NewClient(base db.Queryable) policymodel.Gate {
	return &ClientGate{
		Base: base,
	}
}
