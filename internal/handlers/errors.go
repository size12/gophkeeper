package handlers

import "errors"

var (
	ErrFieldIsEmpty = errors.New("login or password is empty")
	ErrDataIsEmpty  = errors.New("data is empty")
)
