package gate

import (
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/gate/roles"
	"linhazero.com/pvapi/models/policymodel"
)

type AdminGate struct{}

func (admin *AdminGate) Dispatch(
	u User,
	r Resource,
) (bool, error) {
	// admin resources are always local
	return u.Role == roles.Admin, nil
}

func NewAdmin(_ db.Queryable) policymodel.Gate {
	return &AdminGate{}
}
