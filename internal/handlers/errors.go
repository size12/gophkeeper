package handlers

import "errors"

var (
	ErrFieldIsEmpty  = errors.New("field is empty")
	ErrDataIsEmpty   = errors.New("data is empty")
	ErrDataIsInvalid = errors.New("invalid data")
)
