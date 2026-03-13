package main

const (
	// Auth
	userPasswordQuery = `
Select id, password From users Where email = $1
`
)
