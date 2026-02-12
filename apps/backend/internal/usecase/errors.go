package usecase

import "errors"

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrNameRequired    = errors.New("name is required")
	ErrEmailExists     = errors.New("email already registered")
)
