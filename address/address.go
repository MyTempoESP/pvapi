package address

import "github.com/jackc/pgx/v5"

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
