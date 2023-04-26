package handlers

import "errors"

var (
	ErrFieldIsEmpty   = errors.New("field is empty")
	ErrWrongMasterKey = errors.New("wrong master key")
)
