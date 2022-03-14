package repositories

import (
	"github.com/romm80/shortener.git/internal/app"
)

type Shortener interface {
	Add(urlsID []app.URLsID, userID uint64) error
	Get(id string) (string, error)
	GetUserURLs(userID uint64) ([]app.UserURLs, error)
	NewUser() (uint64, error)
	CheckUserID(userID uint64) (bool, error)
	Ping() error
}
