package util

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func HashPassword(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword(
		[]byte(pass), 12,
	)
}

func ComparePasswordWithHash(pass string, hashed []byte) error {
	return bcrypt.CompareHashAndPassword(hashed, []byte(pass))
}

func GenerateRandomPassword(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordCharset))))
		if err != nil {
			return "", err
		}
		b[i] = passwordCharset[n.Int64()]
	}
	return string(b), nil
}
