package userutils

import (
	"context"
	"fmt"

	"github.com/jackc/pgerrcode"
	"linhazero.com/pvapi/db"
	"linhazero.com/pvapi/models/usermodel"
	"linhazero.com/pvapi/util"
)

// User is an incomplete user awaiting insertion
type User struct {
	Name         string
	Email        string
	PasswordHash []byte
}

func JustNew(
	ctx context.Context,
	base db.Queryable,
	name, email, password string,
) error {
	user, err := NewUser(name, email, password)

	if err != nil {
		return err
	}

	_, err = user.Save(ctx, base)

	return err
}

// NewUser builds a noncommited user model instance to be committed to the
// db later. The main purpose is to make a friendly API to hash the password
// secretly and 'Save' the user later when we have a db connection.
func NewUser(name, email, password string) (User, error) {
	var zero User

	// most of the time this function is trying to
	// create a shop anyways, so this should be fine,
	// still TODO/FIXME to avoid performance issues later.
	passwordHash, err := util.HashPassword(password)

	if err != nil {
		return zero, err
	}

	return User{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}, nil
}

// Save inserts the user to the database, or just selects one if already exists
func (u User) Save(
	ctx context.Context,
	base db.Queryable,
) (user int, err error) {
	user, err = usermodel.FindOrCreate(
		ctx, base,
		u.Name, u.Email,
		u.PasswordHash,
	)

	return
}

// JustSave simply saves the user without returning anything but an error on failure
func (u User) JustSave(
	ctx context.Context,
	base db.Queryable,
) error {
	_, err := u.Save(ctx, base)
	return err
}

// MustSave inserts the user to the database, throwing an error if already exists
func (u User) MustSave(
	ctx context.Context,
	base db.Queryable,
) (int, error) {
	user, err := usermodel.Create(
		ctx, base,
		u.Name, u.Email,
		u.PasswordHash,
	)

	if err != nil {
		pge, isPgErr := util.GetPgError(err)
		if isPgErr {
			if pge.Code == pgerrcode.UniqueViolation {
				return 0, fmt.Errorf("user already exists: %w", err)
			}
		}

		return 0, nil
	}

	return user, nil
}
