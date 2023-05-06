package handlers

import "errors"

// Errors for handlers.
var (
	ErrFieldIsEmpty   = errors.New("field is empty")
	ErrWrongMasterKey = errors.New("wrong master key")
)
