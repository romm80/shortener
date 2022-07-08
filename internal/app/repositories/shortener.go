// Package repositories implements an interface for working with the repository
package repositories

import (
	"errors"

	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories/dbpostgres"
	"github.com/romm80/shortener.git/internal/app/repositories/linkedliststorage"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
	"github.com/romm80/shortener.git/internal/app/server"
)

// Shortener repository interface
type Shortener interface {
	Add(url string, userID uint64) (string, error)                                      // adds a link by user id
	AddBatch(urls []models.RequestBatch, userID uint64) ([]models.ResponseBatch, error) // adds a batch of links by user id
	Get(id string) (string, error)                                                      // returns original link by id
	GetUserURLs(userID uint64) ([]models.UserURLs, error)                               // returns user shortened links
	NewUser() (uint64, error)                                                           // adds a new user
	Ping() error                                                                        // database connection check
	DeleteBatch(uint64, []string) error                                                 // batch deleting links by user id
}

// NewStorage returns an initialized database connection
func NewStorage() (Shortener, error) {
	var err error
	var storage Shortener

	switch server.Cfg.DBType {
	case server.DBMap:
		storage, err = mapstorage.New()
	case server.DBPostgres:
		storage, err = dbpostgres.New()
	case server.DBLinkedList:
		storage = linkedliststorage.New()
	default:
		return nil, errors.New("wrong DB type")
	}
	if err != nil {
		return nil, err
	}

	return storage, nil
}
