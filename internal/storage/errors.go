package storage

import "errors"

// Errors for DB storage.
var (
	ErrUserUnauthorized = errors.New("user is unauthorized")
	ErrWrongCredentials = errors.New("wrong login or password")
	ErrLoginExists      = errors.New("this login already exists")
	ErrNotFound         = errors.New("not found record with such id")
	ErrUnknown          = errors.New("internal server error")
)
