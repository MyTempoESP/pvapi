package address

type Address struct {
	Street     string `json:"street" db:"addrstreet"`
	Number     string `json:"number" db:"addrnumber"`
	Neigh      string `json:"neighborhood" db:"addrneigh"`
	City       string `json:"city" db:"addrcity"`
	State      string `json:"state" db:"addrstate"`
	CEP        string `json:"cep" db:"addrcep"`
	Complement string `json:"complement" db:"addrcomplement"`
}
