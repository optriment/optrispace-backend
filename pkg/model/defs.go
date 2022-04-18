package model

import "errors"

// Errors
var (
	ErrEntityNotFound          = errors.New("entity not found")
	ErrValueIsRequired         = errors.New("value is required")
	ErrConnectionAlreadyExists = errors.New("connection already exists")
	ErrUnauthorized            = errors.New("user not authorized")
	ErrDuplication             = errors.New("duplication")
)