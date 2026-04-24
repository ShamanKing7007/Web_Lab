package apperrors

import "errors"

var (
	ErrNotFound   = errors.New("note not found")
	ErrInvalidID  = errors.New("invalid note id")
	ErrValidation = errors.New("validation error")
	ErrForbidden  = errors.New("you do not have permission to perform this action")
)
