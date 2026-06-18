package address

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrPartialAddress = errors.New("endereço parcial: preencha todos os campos obrigatórios ou nenhum")

type Address struct {
	Street     string `json:"street" db:"addrstreet"`
	Number     string `json:"number" db:"addrnumber"`
	Neigh      string `json:"neighborhood" db:"addrneigh"`
	City       string `json:"city" db:"addrcity"`
	State      string `json:"state" db:"addrstate"`
	CEP        string `json:"cep" db:"addrcep"`
	Complement string `json:"complement" db:"addrcomplement"`
}

type AddressRequest struct {
	Street       string `json:"street" binding:"required"`
	Number       string `json:"number" binding:"required"`
	Neighborhood string `json:"neighborhood" binding:"required"`
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	CEP          string `json:"CEP" binding:"required"`
	Complement   string `json:"complement"`
}

func (a Address) Params() pgx.NamedArgs {
	return pgx.NamedArgs{
		"addrstreet":     a.Street,
		"addrnumber":     a.Number,
		"addrneigh":      a.Neigh,
		"addrcity":       a.City,
		"addrstate":      a.State,
		"addrcep":        a.CEP,
		"addrcomplement": a.Complement,
	}
}

type AddressUpdateRequest struct {
	Street       *string `json:"street"`
	Number       *string `json:"number"`
	Neighborhood *string `json:"neighborhood"`
	City         *string `json:"city"`
	State        *string `json:"state"`
	CEP          *string `json:"cep"`
	Complement   *string `json:"complement"`
}

type AddressRequestOptional struct {
	Street       string `json:"street"`
	Number       string `json:"number"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	CEP          string `json:"CEP"`
	Complement   string `json:"complement"`
}

func (a AddressRequest) Prepare() Address {
	return Address{
		Street:     a.Street,
		Number:     a.Number,
		Neigh:      a.Neighborhood,
		City:       a.City,
		State:      a.State,
		CEP:        a.CEP,
		Complement: a.Complement,
	}
}

func (a AddressRequestOptional) Prepare() (Address, error) {
	empty := a.Street == "" && a.Number == "" && a.Neighborhood == "" &&
		a.City == "" && a.State == "" && a.CEP == "" && a.Complement == ""

	if empty {
		return Address{Number: "0"}, nil
	}

	// all-or-nothing: if not empty, all required fields must be filled
	if a.Street == "" || a.Number == "" || a.Neighborhood == "" ||
		a.City == "" || a.State == "" || a.CEP == "" {
		return Address{}, ErrPartialAddress
	}

	return Address{
		Street:     a.Street,
		Number:     a.Number,
		Neigh:      a.Neighborhood,
		City:       a.City,
		State:      a.State,
		CEP:        a.CEP,
		Complement: a.Complement,
	}, nil
}

func (a Address) IsEmpty() bool {
	return a.Number == "0" && a.Street == "" && a.Neigh == "" &&
		a.City == "" && a.State == "" && a.CEP == "" && a.Complement == ""
}

func (a Address) ToResource() *Address {
	if a.IsEmpty() {
		return nil
	}
	return &a
}
