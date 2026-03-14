package query

const (
	// Auth
	UserPasswordQuery = `
Select id, password_hash From users Where email = $1
`
	UserRegisterQuery = `
Insert Into users (name, email, password_hash) Values ($1, $2, $3)
`

	// general
	UserDataQuery = `
Select name, email From users Where id = $1
`
)
