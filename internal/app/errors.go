package app

import "errors"

var (
	ErrConflictURLID = errors.New("conflict url id")
	ErrEmptyRequest  = errors.New("empty request")
	ErrDeletedURL    = errors.New("url deleted")
)
