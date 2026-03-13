package main

const (
	// Auth
	userPasswordQuery = `
Select id, password_hash From users Where email = $1
`
	userRegisterQuery = `
Insert Into users (email, password_hash) Values ($1, $2)
`
)
