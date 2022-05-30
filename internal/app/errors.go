package app

import (
	"errors"
	"log"
	"net/http"
)

var (
	ErrConflictURLID = errors.New("conflict url id")
	ErrEmptyRequest  = errors.New("empty request")
	ErrDeletedURL    = errors.New("url deleted")
)

func ErrStatusCode(err error) int {
	log.Println(err)

	switch {
	case errors.Is(err, ErrConflictURLID):
		return http.StatusConflict
	case errors.Is(err, ErrEmptyRequest):
		return http.StatusBadRequest
	case errors.Is(err, ErrDeletedURL):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}
