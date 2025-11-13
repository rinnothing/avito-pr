package model

import "errors"

var (
	ErrNotFound       = errors.New("item not found")
	ErrAlreadyExists  = errors.New("item already exists")
	ErrWrongInvariant = errors.New("broken invariant")
)
