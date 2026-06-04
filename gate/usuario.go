package gate

import "linhazero.com/pvapi/models/policymodel"

type UsuarioGate struct{}

func (g *UsuarioGate) Dispatch(u User, r Resource) (bool, error) {
	return true, nil
}

func NewUsuario() policymodel.Gate {
	return &UsuarioGate{}
}
