package model

import "errors"

var (
	ErrEntityNotFound          = errors.New("entity not found")
	ErrRequiredFieldNotFilled  = errors.New("required field not filled")
	ErrConnectionAlreadyExists = errors.New("connection already exists")
)
