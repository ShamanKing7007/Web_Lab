package errors

import "errors"

var (
	ErrNotFound      = errors.New("entity not found")
	ErrInvalidID     = errors.New("invalid entity id")
	ErrAlreadyExists = errors.New("entity already exists")
	ErrValidation    = errors.New("validation error")
)
