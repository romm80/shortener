package repositories

import "github.com/romm80/shortener.git/internal/app/repositories/mapstorage"

type Shortener interface {
	Add(string, uint64) (string, error)
	Get(string) (string, error)
	GetUserURLs(uint64) []mapstorage.UserURLs
	NewUser() uint64
	CheckUserID(uint64) bool
}
