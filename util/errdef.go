package util

import (
	"errors"
	"fmt"
)

var (
	ErrBadRequest         = errors.New("malformed request")
	ErrDuplicatedResource = errors.New("trying to create duplicated resource")
	ErrResourceLocked     = errors.New("resource is locked and cannot be modified")
	ErrNotFound           = errors.New("resource not found")
	ErrNotImplemented     = errors.New("function not yet implemented")
)

func NotImplemented(operation string) error {
	return fmt.Errorf(
		"could not fulfill request for operation %q: %w",
		operation, ErrNotImplemented,
	)
}
