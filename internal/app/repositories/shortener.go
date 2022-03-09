package repositories

import "github.com/romm80/shortener.git/internal/app/repositories/mapstorage"

type Shortener interface {
	Add(string, uint64) (string, error)
	Get(string) (string, error)
	NewUser() uint64
	GetUserURLs(uint64) []mapstorage.UserURLs
}
