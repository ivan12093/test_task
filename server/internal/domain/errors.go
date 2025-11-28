package domain

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrNotValidEmail     = errors.New("email not valid")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserNotExists     = errors.New("user not exists")
	ErrInvalidGoogleCode = errors.New("invalid Google code")
)

var (
	ErrSessionNotFound = errors.New("session not found")
)
