package errors

import (
	"errors"
	"fmt"
)

var (
	ErrVaultLocked       = errors.New("vault is locked or not initialized")
	ErrVaultNotExists    = errors.New("vault does not exist")
	ErrVaultExists       = errors.New("vault already exists")
	ErrInvalidPassword   = errors.New("invalid master password")
	ErrAccountNotFound   = errors.New("account not found")
	ErrDataCorrupted     = errors.New("data is corrupted")
	ErrEmptyCharset      = errors.New("at least one character set must be selected")
	ErrInvalidLength     = errors.New("password length must be positive")
	ErrDirectoryRequired = errors.New("data directory cannot be empty")
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
