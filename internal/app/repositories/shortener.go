package repositories

import (
	"errors"
	"github.com/romm80/shortener.git/internal/app/models"
	"github.com/romm80/shortener.git/internal/app/repositories/dbpostgres"
	"github.com/romm80/shortener.git/internal/app/repositories/mapstorage"
	"github.com/romm80/shortener.git/internal/app/server"
)

type Shortener interface {
	Add(url string, userID uint64) (string, error)
	AddBatch(urls []models.RequestBatch, userID uint64) ([]models.ResponseBatch, error)
	Get(id string) (string, error)
	GetUserURLs(userID uint64) ([]models.UserURLs, error)
	NewUser() (uint64, error)
	Ping() error
	DeleteBatch(uint64, []string) error
}

func NewStorage() (Shortener, error) {
	var err error
	var storage Shortener

	switch server.Cfg.DBType {
	case server.DBMap:
		storage, err = mapstorage.New()
	case server.DBPostgres:
		storage, err = dbpostgres.New()
	default:
		return nil, errors.New("wrong DB type")
	}
	if err != nil {
		return nil, err
	}

	return storage, nil
}
