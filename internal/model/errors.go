package model

import "errors"

var (
	ErrNotFound      = errors.New("item not found")
	ErrAlreadyExists = errors.New("item already exists")
	ErrAlreadyMerged = errors.New("pr already merged")
	ErrNoCandidates  = errors.New("no candidates for reassigning")
	ErrNotReviewer   = errors.New("user is not a reviewer")
)
