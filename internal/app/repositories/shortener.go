// Package repositories implements an interface for working with the repository
package repositories

import (
	"errors"

	"github.com/romm80/shortener.git/internal/app"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories/dbpostgres"
	"github.com/romm80/shortener.git/internal/app/repositories/linkedliststorage"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
)

// Shortener repository interface
type Shortener interface {
	Add(url, urlID string, userID uint64) error           // adds a link by user id
	Get(id string) (string, error)                        // returns original link by id
	GetUserURLs(userID uint64) ([]models.UserURLs, error) // returns user shortened links
	NewUser() (uint64, error)                             // adds a new user
	Ping() error                                          // database connection check
	DeleteBatch(uint64, []string) error                   // batch deleting links by user id
	GetStats() (*models.StatsResponse, error)
}

// NewStorage returns an initialized database connection
func NewStorage() (Shortener, error) {
	var err error
	var storage Shortener

	switch app.Cfg.DBType {
	case app.DBMap:
		storage, err = mapstorage.New()
	case app.DBPostgres:
		storage, err = dbpostgres.New()
	case app.DBLinkedList:
		storage = linkedliststorage.New()
	default:
		return nil, errors.New("wrong DB type")
	}
	if err != nil {
		return nil, err
	}

	return storage, nil
}
